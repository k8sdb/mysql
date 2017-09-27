package controller

import (
	"log"
	"testing"
)

func TestController_create(t *testing.T) {

	ctrl := GetNewController()
	msql := DemoMySQL()
	log.Println("creating!!")
	if err := ctrl.create(msql); err != nil {
		log.Println("error creating MySQL", err)
	} else {
		log.Println("Creating task succesfull")
	}
}
func TestController_Run(t *testing.T) {
	ctrl := GetNewController()
	ctrl.Run()
}
