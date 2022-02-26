package signserver

//revive:disable

type RespID uint16

//go:generate stringer -type=RespID
const (
	SIGN_UNKNOWN RespID = iota
	SIGN_SUCCESS
	SIGN_EFAILED   // Authentication server communication failed
	SIGN_EILLEGAL  // Incorrect input, authentication has been suspended
	SIGN_EALERT    // Authentication server process error
	SIGN_EABORT    // The internal procedure of the authentication server ended abnormally
	SIGN_ERESPONSE // Procedure terminated due to abnormal certification report
	SIGN_EDATABASE // Database connection failed
	SIGN_EABSENCE
	SIGN_ERESIGN
	SIGN_ESUSPEND_D
	SIGN_ELOCK
	SIGN_EPASS
	SIGN_ERIGHT
	SIGN_EAUTH
	SIGN_ESUSPEND   // This account is temporarily suspended. Please contact customer service for details
	SIGN_EELIMINATE // This account is permanently suspended. Please contact customer service for details
	SIGN_ECLOSE
	SIGN_ECLOSE_EX // Login process is congested. <br> Please try to sign in again later
	SIGN_EINTERVAL
	SIGN_EMOVED
	SIGN_ENOTREADY
	SIGN_EALREADY
	SIGN_EIPADDR // Region block because of IP address.
	SIGN_EHANGAME
	SIGN_UPD_ONLY
	SIGN_EMBID
	SIGN_ECOGCODE
	SIGN_ETOKEN
	SIGN_ECOGLINK
	SIGN_EMAINTE
	SIGN_EMAINTE_NOUPDATE

	// Couldn't find names for the following:
	UNK_32
	UNK_33
	UNK_34
	UNK_35

	SIGN_XBRESPONSE
	SIGN_EPSI
	SIGN_EMBID_PSI
)
