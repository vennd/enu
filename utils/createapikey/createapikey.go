package main

import (
	"strings"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	//	"strconv"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/log"
)

type ApiKey struct {
	Comment string
	Key     string
	Secret  string
}

func main() {
	log.Println("Generating new API key")

	reader := bufio.NewReader(os.Stdin)
	log.Printf("Enter comment for key:")
	comment, err := reader.ReadString('\n')

	if err != nil {
		log.Fluentf(consts.LOGERROR, err.Error())
	}

	log.Printf("Enter user id (must be integer) for key:")
	//	userid, err := reader.ReadString('\n')

	var userid int64

	_, err2 := fmt.Scan(&userid)
	if err2 != nil {
		log.Fluentf(consts.LOGERROR, err.Error())
	}

	var confirm string
	log.Printf("Creating an API key\nBlockchain: counterparty\nUserId: %d\nComment: %s\n\nConfirm (Y to confirm, everything else cancels)?", userid, comment)
	_, err3 := fmt.Scan(&confirm)

	if err3 != nil {
		log.Fluentf(consts.LOGERROR, err3.Error())
		return
	}

	if strings.TrimSpace(confirm) != "Y" {
		log.Println(confirm)
		log.Println("Aborted")
		return
	}

	key, secret, err4 := database.CreateUserKey(userid, "", "counterparty", "", "")
	if err4 != nil {
		log.Fluentf(consts.LOGERROR, err4.Error())
		return
	}

	f, err := os.Create("enu_key.json")
	if err != nil {
		log.Fluentf(consts.LOGERROR, err.Error())
		return
	}
	defer f.Close()

	var apikey = ApiKey{Comment: strings.TrimSpace(comment), Key: key, Secret: secret}

	json, err := json.MarshalIndent(apikey, "", "  ")
	if err != nil {
		log.Fluentf(consts.LOGERROR, err.Error())
		return
	}

	log.Println(string(json))

	_, err5 := f.Write(json)
		if err5 != nil {
		log.Fluentf(consts.LOGERROR, err5.Error())
		return
	}

	log.Println("Created user key")

}
