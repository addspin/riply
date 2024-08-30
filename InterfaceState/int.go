package InterfaceState

import (
	"fmt"
	"riply/Check"
	"log"
	"net"
	"sync"
	"time"

	"github.com/vishvananda/netlink"
)

type MutexInterface struct {
	MutexInterface sync.Mutex
}

func (m *MutexInterface) InterfaceUp(interfaceName, ipAddress string) {

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

	// получаем атрибуты интерфейса
	attrs := link.Attrs()
	// проверяем статус интерфейса 0 - down, 1 - up
	if attrs.Flags&net.FlagUp == 0 {
		// устанавливаем статус интерфейса в "up"
		err = netlink.LinkSetUp(link)
		if err != nil {
			log.Fatal(err)
		}
	}

	// если адрес есть, ничего не делаем
	if len(ipCheck) != 0 {
		return
	}
	// елис адреса нет, устанавливаем его
	if len(ipCheck) == 0 {
		err = netlink.AddrAdd(link, addr)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (m *MutexInterface) InterfaceDown(interfaceName, ipAddress string) {

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

	// получаем атрибуты интерфейса
	attrs := link.Attrs()
	// проверяем статус интерфейса 0 - down, 1 - up
	if attrs.Flags&net.FlagUp != 0 {
		// устанавливаем статус интерфейса в "down"
		err = netlink.LinkSetDown(link)
		if err != nil {
			log.Fatal(err)
		}
	}
	// если адреса нет, ничего не делаем
	if len(ipCheck) == 0 {
		return
	}
	// если адрес есть, удаляем его
	if len(ipCheck) != 0 {
		err = netlink.AddrDel(link, addr)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (m *MutexInterface) UpDown(interfaceName, ipAddress string, statusCodeSync *Check.StatusCodeSync, statusCodeTcp *Check.StatusCodeTcp) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	m.MutexInterface.Lock()
	defer m.MutexInterface.Unlock()
	for range ticker.C {
		master := 1
		slave := 2
		none := 0

		// log.Println("Remote Status: ", statusCodeSync.ExitCodeSync)
		if statusCodeSync.ExitCodeSync == master {
			m.InterfaceDown(interfaceName, ipAddress)
		}
		if statusCodeSync.ExitCodeSync == slave {
			m.InterfaceUp(interfaceName, ipAddress)
		}
		if statusCodeSync.ExitCodeSync == none {
			m.InterfaceUp(interfaceName, ipAddress)

		}
	}
}
