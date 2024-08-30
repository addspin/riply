package main

import (
	"fmt"
	"lb/Check"
	"lb/InterfaceState"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/viper"
	"github.com/vishvananda/netlink"
)

func main() {

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	serversSync := viper.GetString("serversSync.srv")
	serverSyncPort := viper.GetString("serversSync.port")
	appPort := viper.GetString("app.port")
	scriptName := viper.GetString("type.sh")
	tcpPort := viper.GetString("type.tcp.port")
	tcpHost := viper.GetString("type.tcp.host")
	interfaceName := viper.GetString("interfaces.name")
	ipAddress := viper.GetString("interfaces.ip")

	log.Println("serversSync: ", serversSync)
	log.Println("scriptName: ", scriptName)
	log.Println("tcpPort: ", tcpPort)
	log.Println("tcpHost: ", tcpHost)

	//---------------------------------------Create a new engine Template
	// engine := html.New("./template", ".html")
	//---------------------------------------Pass the engine to the Views
	app := fiber.New(fiber.Config{
		CaseSensitive: false,
		// StrictRouting: true,
		ServerHeader:     "RIP",
		AppName:          "App v1.0.0",
		DisableKeepalive: true,

	})

	// Очищаем все интерфейсы
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		fmt.Println(err)
	}
	ipCheck, err := iface.Addrs()
	if err != nil {
		fmt.Println(err)
	}
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		fmt.Println(err)
	}
	addr, err := netlink.ParseAddr(ipAddress)
	if err != nil {
		fmt.Println(err)
	}
	if len(ipCheck) != 0 {
		err = netlink.AddrDel(link, addr)
		if err != nil {
			fmt.Println(err)
		}
	}

	if scriptName != "" {
		timeCheckSh := 2
		statusCodeSh := &Check.StatusCodeSh{}
		go statusCodeSh.Sh(scriptName, timeCheckSh)

		app.Get("/sh", func(c fiber.Ctx) error {
			if c.Method() == "GET" {
				statusCodeSh.MutexSh.Lock()
				defer statusCodeSh.MutexSh.Unlock()
				statusStrSh := strconv.Itoa(statusCodeSh.ExitCodeSh)
				statusStr := "Error code: " + statusStrSh
				return c.SendString(statusStr)
			}
			return nil
		})

		app.Get("/state", func(c fiber.Ctx) error {
			if c.Method() == "GET" {
				statusCodeSh.MutexSh.Lock()
				defer statusCodeSh.MutexSh.Unlock()
				statusStrSh := statusCodeSh.ExitCodeSh
				if statusStrSh == 1 {
					statusStr := "Slave" // if not available
					return c.SendString(statusStr)
				}
				if statusStrSh == 0 {
					statusStr := "Master" // if available
					return c.SendString(statusStr)
				}
			}
			return nil
		})
	}

	if tcpHost != "" && tcpPort != "" {
		statusCodeSync := &Check.StatusCodeSync{}
		statusCodeTcp := &Check.StatusCodeTcp{}
		InterfaceState := &InterfaceState.MutexInterface{}

		// Проверяем статус уделнного клиена
		go statusCodeSync.Sync(serversSync, serverSyncPort)
		// Проверяем запущен отслеживаемый сервис или нет
		time.Sleep(2 * time.Second)
		go statusCodeTcp.TCPPortAvailable(tcpHost, tcpPort, statusCodeSync)

		// Постоянно проверяем я slave или master, если master поднимаем интерфейс.
		go InterfaceState.UpDown(interfaceName, ipAddress, statusCodeSync, statusCodeTcp)
		// Если статус не равно Master устанавливает state в Master

		// app.Get("/tcp_state", func(c fiber.Ctx) error {
		// 	if c.Method() == "GET" {
		// 		statusCodeTcp.MutexTcp.Lock()
		// 		defer statusCodeTcp.MutexTcp.Unlock()
		// 		statusStrTcp := statusCodeTcp.ExitCodeTcp
		// 		if !statusStrTcp {
		// 			// tcpPortStr := strconv.Itoa(tcpPort)
		// 			statusStr := "unavailable"
		// 			return c.SendString(statusStr)
		// 		}
		// 		if statusStrTcp {
		// 			// tcpPortStr := strconv.Itoa(tcpPort)
		// 			statusStr := "available"
		// 			return c.SendString(statusStr)
		// 		}
		// 	}
		// 	return nil
		// })
		// Мой статус
		app.Get("/state", func(c fiber.Ctx) error {
			if c.Method() == "GET" {

				myStateTcp := statusCodeTcp.MyState

				if myStateTcp == "Master" {
					log.Println("My Status: Master")
					statusStr := "Master" // Если удаленный клиент не Master становимся Master
					return c.SendString(statusStr)

				}
				if myStateTcp == "Slave" {
					log.Println("My Status: Slave")
					statusStr := "Slave" // Если удаленный клиент Master становимся Slave
					return c.SendString(statusStr)
				}
			}
			return nil
		})
	}
	log.Fatal(app.Listen(":" + appPort, fiber.ListenConfig{
    // EnablePrefork: true,
	EnablePrintRoutes: false,
    DisableStartupMessage: false},))
}
