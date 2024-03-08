package main

/*
#cgo LDFLAGS: -lpam
#include <security/pam_ext.h>
#include <security/pam_modules.h>
#include <security/pam_appl.h>
#include <pwd.h>

extern int open_session(pam_handle_t *pamh);

const char* c_username;
const char* c_password;
int uid;
int gid;

int get_authtok(pam_handle_t* pamh) {
    return pam_get_authtok(pamh, PAM_AUTHTOK, &c_password , NULL);
}

int get_user(pam_handle_t* pamh) {
    return pam_get_user(pamh, &c_username, "Username: ");
}

int pam_sm_setcred(pam_handle_t *pamh, int flags, int argc, const char **argv) {
    return PAM_SUCCESS;
}

int pam_sm_open_session(pam_handle_t *pamh, int flags, int argc, const char **argv) {
	return open_session(pamh);
}

int pam_sm_close_session(pam_handle_t *pamh, int flags, int argc, const char **argv) {
	return PAM_SUCCESS;
}

int get_ids(pam_handle_t* pamh) {
	const char* username;
	int res = pam_get_user(pamh, &username, NULL);
	if (res != PAM_SUCCESS)
		return res;

	// Fetch passwd entry 
	const struct passwd *pwent = getpwnam(username);
	if (!pwent)
	{
		pam_error(pamh, "User not found in passwd");
		return PAM_CRED_INSUFFICIENT;
	}
	uid = pwent->pw_uid;
	gid = pwent->pw_gid;

	return PAM_SUCCESS;
}
*/
import "C"

import (
	"errors"
	"log"
)

const NOBODY_ID = 65534

func GetUser(pamh *C.pam_handle_t) (string, error) {
	ret := C.get_user(pamh)
	if ret != C.PAM_SUCCESS {
		log.Println("Username could not be retrieved")
		return "", errors.New("username could not be retrieved")
	}
	return C.GoString(C.c_username), nil
}

func get_uid_and_gid(pamh *C.pam_handle_t) (int, int, error) {
	ret := C.get_ids(pamh)
	if ret != C.PAM_SUCCESS {
		log.Println("User ID could not be retrieved")
		return NOBODY_ID, NOBODY_ID, errors.New("User ID could not be retrieved")
	}
	return int(C.uid), int(C.gid), nil
}

func GetPassword(pamh *C.pam_handle_t) (string, error) {
	ret := C.get_authtok(pamh)
	if ret != C.PAM_SUCCESS {
		log.Println("User password could not be retrieved")
		return "", errors.New("user password could not be retrieved")
	}
	return C.GoString(C.c_password), nil
}