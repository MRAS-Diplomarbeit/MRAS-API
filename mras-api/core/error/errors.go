package error

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	AUTH001 = Error{Code: "AUTH001", Message: "Missing or Invalid Authorization Header"}
	AUTH002 = Error{Code: "AUTH002", Message: "JWT not found in Redis (Expired)"}
	AUTH003 = Error{Code: "AUTH003", Message: "User not Found in Database (Wrong Username or Password)"}
	AUTH004 = Error{Code: "AUTH004", Message: "User already exists"}
	AUTH005 = Error{Code: "AUTH005", Message: "Invalid RefreshToken"}
	AUTH006 = Error{Code: "AUTH006", Message: "User not found"}
	AUTH007 = Error{Code: "AUTH007", Message: "Wrong Reset Code"}
	AUTH008 = Error{Code: "AUTH008", Message: "Password not Reset"}
	AUTH009 = Error{Code: "AUTH009", Message: "User not Authorized for this Action"}

	DBSQ001 = Error{Code: "DBSQ001", Message: "Error Accessing Database"}
	DBSQ002 = Error{Code: "DBSQ002", Message: "Error Saving RefreshToken in Database"}
	DBSQ003 = Error{Code: "DBSQ003", Message: "Error Accessing Redis"}
	DBSQ004 = Error{Code: "DBSQ004", Message: "Error Reseting User Password"}
	DBSQ005 = Error{Code: "DBSQ005", Message: "Error Saving new Password"}
	DBSQ006 = Error{Code: "DBSQ006", Message: "User ID not Found"}
	DBSQ007 = Error{Code: "DBSQ007", Message: "Error Saving Speaker"}
	DBSQ008 = Error{Code: "DBSQ008", Message: "Speaker ID not Found"}

	RQST001 = Error{Code: "RQST001", Message: "Error decoding Request"}
	RQST002 = Error{Code: "RQST002", Message: "Request Body missing fields"}
	CLIE001 = Error{Code: "CLIE001", Message: "Speaker/s not active"}
	CLIE002 = Error{Code: "CLIE002", Message: "Error sending Request to client"}
	CLIE003 = Error{Code: "CLIE003", Message: "Error creating Session. Aborting playback!"}
	CLIE004 = Error{Code: "CLIE004", Message: "Error decoding Client Response"}
)
