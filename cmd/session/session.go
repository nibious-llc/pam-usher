package main

/*
#cgo LDFLAGS: -lpam
#include <security/pam_appl.h>
#include <security/pam_modules.h>
*/
import "C"

import (
	"log"
	"gopkg.in/yaml.v2"
	"os"
	"io/ioutil"
	"nibious.com/pam_usher/internal/pam_usher"
	"errors"
	"io/fs"
)


// Read a configuration file written in yaml. The directory is hard coded and
// loads the configuration file directly into a structure.
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

// Creates directories based on the configuration file. The base directory
// specified will be owned by root with 0755 permissions. The user specific
// directory will have the permissions set properly with 0700. The user is free
// to change this later and this function will not override that. This function
// may look a little odd, but it is to reduce the amount of system calls
// required and thus speed up the session creation. 
//
// First, we check if the complete directory is already existing. If it is, we
// continue. If it isn't, we create the parent directories. If that fails, we
// return a failure. Otherwise, we go to the beginning where we were going to
// create our complete directory again and try it. 
//
// A `goto` may have been an odd choice, but it enables a clean jump back to our
// pre-opimization of reducing system calls required while creating various
// directories. To prevent loops, we do check if the error is something besides
// an 'ErrExist', and if so, we then return an error.
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

create_dir:
		// Check if the directory exists by trying to create it. If it does
		// exist, we will get an error, if it doesn't error (i.e. makes the
		// directory) then we will get what we need done.
		if err := os.Mkdir(user_home_folder, 0700); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// It likely failed because the parent directory doesn't exist.
				// Let's create it.
				if err := os.MkdirAll(v, 0755); err != nil {
					if !os.IsExist(err) {
						log.Println("error: Could not create main directory: %v", err)
						return err
					}
				} else {
					// Change the owner to root on our new folder.
					if err := os.Chown(v, 0, 0); err != nil {
						log.Println("error: Could not change ownership of main directory: %v", err)
						return err
					}

					// Let's try this create function again
					goto create_dir
				}
			} else if !errors.Is(err, fs.ErrExist) {
				log.Println("error: Failed to create user home folder: ", err)
				return err
			} else {
				continue
			}
		} else {
			// We are here because os.Mkdir succeeded and created a directory. 
			// Now let's change the owner to the user signing in
			if err := os.Chown(user_home_folder, uid, gid); err != nil {
				log.Println("error: %v", err)
				return err
			}
		}
	}
	return nil
}

//export open_session
func open_session(pamh *C.pam_handle_t) C.int {

	// Read the configuration file first
	config, err := read_config()
	if err != nil {
		return C.PAM_SESSION_ERR
	}

	if create_directories(pamh, config) != nil {
		return C.PAM_SESSION_ERR
	}

	return C.PAM_SUCCESS
}

// This is required for a shared object to be compiled properly with CGO
func main() {}