package store

import (
	"testing"
)

func TestCheckDuplicateServer(t *testing.T) {
	existingServers := []Server{
		{
			HostName: "server1.example.com",
			IP:       "192.168.1.10",
			Alias:    "srv1",
			User:     "root",
		},
		{
			HostName: "server2.example.com",
			IP:       "192.168.1.20",
			Alias:    "srv2",
			User:     "admin",
		},
	}

	// Test exact duplicate
	dupServer := Server{
		HostName: "server1.example.com",
		IP:       "192.168.1.10",
		Alias:    "srv1",
		User:     "root",
	}
	if !checkDuplicateServer(dupServer, existingServers) {
		t.Errorf("expected duplicate detection for identical server")
	}

	// Test same alias different IP
	aliasDup := Server{
		HostName: "other.example.com",
		IP:       "192.168.1.30",
		Alias:    "srv1",
		User:     "root",
	}
	if !checkDuplicateServer(aliasDup, existingServers) {
		t.Errorf("expected duplicate detection for identical alias")
	}

	// Test unique server
	uniqueServer := Server{
		HostName: "server3.example.com",
		IP:       "192.168.1.30",
		Alias:    "srv3",
		User:     "root",
	}
	if checkDuplicateServer(uniqueServer, existingServers) {
		t.Errorf("expected unique server to not be marked as duplicate")
	}
}
