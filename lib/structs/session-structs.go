package structs

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Structs                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// SessionInfo holds the information about the federated session.
type SessionInfo struct {
	AccountName    string
	AccountNumber  string
	AccountTypeID  uint
	AwsIamRoleName string
	Region         string
}
