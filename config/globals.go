package config

const (
	AppName   = "s3stress"
	AppNameUC = "S3STRESS"
)

var (
	GlobalQuiet   = false // Quiet flag set via command line
	GlobalJSON    = false // Json flag set via command line
	GlobalDebug   = false // Debug flag set via command line
	GlobalNoColor = false // No Color flag set via command line
)
