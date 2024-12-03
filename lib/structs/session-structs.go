package structs

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Structs                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

type SessionInfo struct {
	AccountName    string
	AccountNumber  string
	AccountTypeID  uint
	AwsIamRoleName string
	Region         string
}
