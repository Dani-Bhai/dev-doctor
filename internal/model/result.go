package model

type Status int

const (
		StatusOk Status = iota
	  StatusWarning
	  StatusError
)

type Result struct {
	  Name string
	 	Group string 
		Status Status
		Message string 
		Details []string
}
