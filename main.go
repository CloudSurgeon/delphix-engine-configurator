package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	deHostname := flag.String("hostname", "", "the hostname or IP address of the Delphix Engine")
	newSysPass := flag.String("syspass", "", "the new sysadmin password")
	newAdminPass := flag.String("dapass", "", "the new delphix_admin password")
	keyFileUsage := "Setting this flag will write the Delphix Engine sshPublicKey to the specified file. Contents are overwritten."
	keyFileName := flag.String("filename", "", keyFileUsage)
	flag.Parse()

	//todo: add OS env variables

	deURL := "http://" + *deHostname
	sysName := "sysadmin"
	initialSysPass := "sysadmin"
	adminName := "delphix_admin"

	sysConfig := Config{
		url:      deURL + "/resources/json/delphix",
		username: sysName,
		password: initialSysPass,
	}

	adminConfig := Config{
		url:      deURL + "/resources/json/delphix",
		username: adminName,
		password: *newAdminPass,
	}

	log.Println("[INFO] Initializing Delphix client for sysadmin")

	sysClient := *sysConfig.Client()
	if err := sysClient.WaitForEngineReady(10, 600); err != nil {
		log.Fatalf("Failed to authenticate. Has this engine been previously configured? ERROR: %s", err)
		os.Exit(1)
	}
	if err := sysClient.LoadAndValidate(); err != nil {
		log.Fatalf("ERROR:", err)
		os.Exit(1)
	}

	if k, err := sysClient.ReturnSshPublicKey(); err != nil {
		log.Fatalf("Failed to get SshPublicKey: %s\n", err)
		os.Exit(1)
	} else {
		log.Printf("[INFO] SSH Public Key: %s\n", k)
		if *keyFileName != "" {
			ioutil.WriteFile(*keyFileName, []byte(k), 0666)
			log.Printf("[INFO] Wrote key to %s\n", *keyFileName)
		}
	}

	if err := sysClient.UpdateUserPasswordByName(sysName, *newSysPass); err != nil {
		log.Fatalf("Failed to update sysadmin password: %s\n", err)
		os.Exit(1)
	}

	if _, err := sysClient.InitializeSystem(adminName, *newAdminPass); err != nil {
		log.Fatalf("Failed to create storage domain. Either this Delphix Engine has already been configured, or it has no storage devices. Exiting.: %s\n", err)
		os.Exit(1)
	}

	log.Println("[INFO] Initializing Delphix client for delphix_admin")

	adminClient := *adminConfig.Client()

	if err := adminClient.LoadAndValidate(); err != nil {
		log.Fatalf("ERROR:", err)
		os.Exit(1)
	}

	if err := adminClient.UpdateUserPasswordByName(adminName, *newAdminPass); err != nil {
		log.Fatalf("Failed to update delphix_admin password: %s\n", err)
		os.Exit(1)
	}
}
