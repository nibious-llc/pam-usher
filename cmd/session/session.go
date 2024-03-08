package main

/*
#cgo LDFLAGS: -lpam
#include <security/pam_appl.h>
#include <security/pam_modules.h>
*/
import "C"

import (
	"errors"
	"log"
	"gopkg.in/yaml.v2"
	"os"
	"io/ioutil"
	"nibious.com/pam_usher/internal/pam_usher"
)

func read_config() (pam_usher.Config, error) {
	c := pam_usher.Config{}

	data, err := ioutil.ReadFile("/etc/nibious/config.yaml")
	if err != nil {
		log.Println("error: %v", err)
		return c, err
	}

    if err := yaml.Unmarshal(data, &c); err != nil {
        log.Println("error: %v", err)
		return c, err
    }

	return c, nil
}

func create_directories(pamh *C.pam_handle_t, config pam_usher.Config) (error) {

	uid, gid, err := get_uid_and_gid(pamh)
	if err != nil {
		return err
	}

	username, err1 := GetUser(pamh)
	if err1 != nil {
		return err
	}

	for _, v := range config.UserDirectories {

		user_home_folder := v + "/" + username

		// Check for directory existance. If exists, let's continue
		d, err := os.Stat(user_home_folder)
		if err == nil && d.IsDir() {
			// Continue making other directories because this exists
			continue
		} else if err != nil && errors.Is(err, os.ErrNotExist) {
			// Ignore the rest of this if statement
		} else if err != nil && !errors.Is(err, os.ErrNotExist)  {
			// We have a real error that needs to be dealt with
			log.Println("error: stat returned an error: %v", err)
			return err
    	} else if d != nil && !d.IsDir() {
			log.Println("error: requested user directory is not a directory.")
			continue
		}

		if d, err := os.Stat(v); err != nil || !d.IsDir() {
			if errors.Is(err, os.ErrNotExist) {
				// Create the root directory if required
				if err := os.Mkdir(v, 0755); err != nil {
					log.Println("error: Could not create main directory: %v", err)
					return err
				}

				// Change the owner to root
				if err := os.Chown(v, 0, 0); err != nil {
					log.Println("error: Could not change ownership of main directory: %v", err)
					return err
				}
			} else if !d.IsDir() {
				log.Println("error: requested base directory is not a directory")
			} else {
				log.Println("error: Unknown error happened: %v", err)
				return err
			}
    	}
		
		// Create the user folder inside
		if err := os.Mkdir(user_home_folder, 0700); err != nil {
			log.Println("error: %v", err)
			return err
		}

		// Change the owner to the user signing in
		if err := os.Chown(user_home_folder, uid, gid); err != nil {
			log.Println("error: %v", err)
			return err
		}
	}
	return nil
}

//export open_session
func open_session(pamh *C.pam_handle_t) C.int {

	// Read the configuration file first
	/*config, err := read_config()
	if err != nil {
		return C.PAM_SESSION_ERR
	}*/

	config := pam_usher.Config{
		UserDirectories: []string{"/tmp/users"},
	}

	if create_directories(pamh, config) != nil {
		return C.PAM_SESSION_ERR
	}

	return C.PAM_SUCCESS
}


func main() {}