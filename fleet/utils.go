package fleet

import (
	"bufio"
	"os"
	"strings"
)

const (
	coreosHostIPs       = "/etc/environment"
	coreosPublicIPv4Var = "COREOS_PUBLIC_IPV4"
)

// CoreosHostPublicIP extracts the CoreOS Public IP from /etc/environment
func CoreosHostPublicIP() (string, error) {
	inFile, err := os.Open(coreosHostIPs)
	if err != nil {
		return "", nil
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, coreosPublicIPv4Var) {
			ip := strings.Replace(line, "COREOS_PUBLIC_IPV4=", "", -1)
			return ip, nil
		}
	}
	return "", nil
}
