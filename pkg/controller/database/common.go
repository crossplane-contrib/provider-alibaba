package database

const (
	ErrNotInstance string = "managed resource is not an instance custom resource"

	ErrNoProvider          = "no provider config or provider specified"
	ErrCreateClient        = "cannot create client"
	ErrGetProvider         = "cannot get provider"
	ErrGetProviderConfig   = "cannot get provider config"
	ErrTrackUsage          = "cannot track provider config usage"
	ErrNoConnectionSecret  = "no connection secret specified"
	ErrGetConnectionSecret = "cannot get connection secret"

	ErrCreateFailed        = "cannot create instance"
	ErrCreateAccountFailed = "cannot create database account"
	ErrDeleteFailed        = "cannot delete instance"
	ErrDescribeFailed      = "cannot describe instance"

	ErrFmtUnsupportedCredSource = "credentials source %q is not currently supported"
)
