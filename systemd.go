package main

import (
	"strings"
	"bytes"
	"os/exec"
	"github.com/coreos/go-systemd/dbus"
)

type ServiceUnit struct {
	File dbus.UnitFile
	Status dbus.UnitStatus
}

// returns all service type units, regardless of their status
func getAllServiceUnits(conn *dbus.Conn) ([]ServiceUnit, error) {
	sdUnitFiles, err := conn.ListUnitFiles()
	if err != nil {
		return nil, err
	}
	sdUnits, err := conn.ListUnits()
	if err != nil {
		return nil, err
	}

	numOfServiceUnits := 0
	for _, unitFile := range sdUnitFiles {
		if strings.HasSuffix(unitFile.Path, ".service") {
			numOfServiceUnits++
		}
	}
	serviceUnits := make([]ServiceUnit, numOfServiceUnits)
	index := 0
	for _, unitFile := range sdUnitFiles {
		if !strings.HasSuffix(unitFile.Path, ".service") {
			continue
		}
		serviceUnits[index].File = unitFile
		strSplit := strings.Split(unitFile.Path, "/")
		serviceName := strSplit[len(strSplit)-1]
		for _, unitStatus := range sdUnits {
			if unitStatus.Name == serviceName {
				serviceUnits[index].Status = unitStatus
				break
			}
		}
		index++
	}

	return serviceUnits, nil
}

// strips the service name from full service file path
func getServiceName(unitPath string) string {
	strSplit := strings.Split(unitPath, "/")
	serviceName := strSplit[len(strSplit)-1]

	return serviceName
}

// obtains service status through "systemctl status"
func getServiceStatus(unitPath string) string {
	cmd := exec.Command("systemctl", "status", getServiceName(unitPath))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run() // the err is not checked because systemctl returns non-zero code

	return out.String()
}
