package utils

import (
	"net"
)

func IsValidMacAddress(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil
}
