package main

import (
	"fmt"
	"time"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/email"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/targets"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/ups"
	"github.com/BrunoTeixeira1996/nutgmonitor/internal/webhook"
)

const version = "2.0"

var upsTargets = targets.InitTargets()

func run() error {
	for _, t := range upsTargets {
		if err := t.ValidateSSHKeys(); err != nil {
			return err
		}
	}
	logger.Log.Printf("[run info] validated all targets\n")

	em, p := email.GetEnvs()
	if em == "" || p == "" {
		e := "[run error] email and password not set for email\n"
		logger.Log.Printf(e)
		return fmt.Errorf(e)
	}

	// FIXME: dont use upsTargets[2] to use nas1 , instead use something like upsTargets["nas1"]
	go ups.AlertFastPowerOff("/app/logs/upslog/upslog.txt", upsTargets[2])
	// for debug
	//go ups.AlertFastPowerOff("upslog.txt", upsTargets[2])

	// waits for updates from Alert Manager
	if err := webhook.StartWebHook(upsTargets); err != nil {
		return err
	}

	// Content field is assigned inside email.buildEmail() function
	finalResult := &email.EmailTemplate{
		Timestamp: time.Now().String(),
	}

	logger.Log.Printf("[run info] preparing email fields\n")
	if err := email.SendEmail(finalResult); err != nil {
		logger.Log.Printf("[run error] could not send email: %s", err)
	}

	return nil
}

func main() {
	err := logger.Setup("logs")
	if err != nil {
		logger.Log.Println(err)
		return
	}

	logger.Log.Println("Running on version: ", version)

	if err := run(); err != nil {
		logger.Log.Println(err)
	}

	logger.Log.Println("Shuting down Pinute ...")
	// Turn off Pinute after all targets are down
	upsTargets[3].ShutdownFunc(upsTargets[3].SSHKey, upsTargets[3].IP)
}

/*
TODO:
   I need to create a new rule in alertmanager to send a telegram notification when the docker container for nutups stops working (DONE)
   Now that i created a new rule in alertmanager and fixed this in this code i need to curl http://192.168.30.13:9995/metrics?target=192.168.30.13:3493 because if that returns "Failed to connect to target: Connection refused (os error 111)" this means that the nut ups docker container got some error and stoped working properly"
*/
