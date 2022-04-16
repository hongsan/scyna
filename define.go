package scyna

const (
	GEN_GET_ID_URL         = "/scyna/generator/id"
	GEN_GET_SN_URL         = "/scyna/generator/sn"
	SESSION_CREATE_URL     = "/scyna/session/create"
	SESSION_UPDATE_CHANNEL = "scyna.session.update"
	SESSION_END_CHANNEL    = "scyna.session.end"
	LOG_WRITE_CHANNEL      = "scyna.log"
	SETTING_WRITE_URL      = "/scyna/setting/write"
	SETTING_READ_URL       = "/scyna/setting/read"
	SETTING_REMOVE_URL     = "/scyna/setting/remove"
	SETTING_UPDATE_CHANNEL = "scyna.setting.updated."
	CALL_WRITE_CHANNEL     = "scyna.call.write"
	APP_UPDATE_CHANNEL     = "scyna.application.updated"
	CLIENT_UPDATE_CHANNEL  = "scyna.client.updated"
	AUTH_CREATE_URL        = "/scyna/auth/create"
	AUTH_GET_URL           = "/scyna/auth/get"
	AUTH_LOGOUT_URL        = "/scyna/auth/logout"
)

var (
	OK                 = &Error{Code: 0, Message: "Success"}
	SERVER_ERROR       = &Error{Code: 1, Message: "Server Error"}
	BAD_REQUEST        = &Error{Code: 2, Message: "Bad Request"}
	PERMISSION_ERROR   = &Error{Code: 4, Message: "Permission Error"}
	REQUEST_INVALID    = &Error{Code: 5, Message: "Request Invalid"}
	MODULE_NOT_EXISTED = &Error{Code: 6, Message: "Module Not Existed"}
)
