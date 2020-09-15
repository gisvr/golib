package ecode

var (
	Chain_Transaction_Invalid_Address        = New(10001)
	Chain_Transaction_Invalid_Parameters     = New(10002)
	Chain_Transaction_Insufficient_Funds     = New(10003)
	Chain_Transaction_Service_Fault          = New(10004)
	Chain_Transaction_Upstream_Service_Fault = New(10005)
	Chain_Transaction_Chain_Service_Fault    = New(10006)
	Chain_Transaction_Invalid_PubKey         = New(10007)
	Chain_Transaction_Unsupport_Chain        = New(10008)
	Chain_Transaction_Broadcast_Fault        = New(10009)
	Chain_Transaction_Unsupport_Token        = New(10010)
	Chain_Transaction_Tx_Unserial_Fault      = New(10011)
	Chain_Transaction_Hex_Decode_Fault       = New(10012)

	//
	Chain_Query_Invalid_Chain      = New(11000)
	Chain_Query_Invalid_Address    = New(11001)
	Chain_Query_Account_NotExist   = New(11002)
	Chain_Query_Rpc_Fault          = New(11003)
	Chain_Query_Mongodb_Fault      = New(11004)
	Chain_Query_Invalid_Parameters = New(11005)
	Chain_Query_Not_Implemented    = New(11006)
	Chain_Query_Unsupport_Token    = New(11007)

	Chain_Index_Invalid_Chain           = New(12000)
	Chain_Index_Mongo_Fault             = New(12001)
	Chain_Index_Subscribe_Address_IsNil = New(12002)
)
