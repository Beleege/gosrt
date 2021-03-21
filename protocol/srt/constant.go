package srt

const (
	HSv5Magic = 0x4A17

	PTypeData    = 0
	PTypeControl = 1

	HSv4 = 4
	HSv5 = 5

	CTHandShake      = 0x0000
	CTKeepalive      = 0x0001
	CTAck            = 0x0002
	CTNAck           = 0x0003
	CTCongestionWarn = 0x0004
	CTShutdown       = 0x0005
	CTAckAck         = 0x0006
	CTDropReq        = 0x0007
	CTPeerErr        = 0x0008
	CTUserDef        = 0x7FFF

	HSTypeWaveHand   = 0x00000000
	HSTypeInduction  = 0x00000001
	HSTypeDone       = 0xFFFFFFFD
	HSTypeAgreement  = 0xFFFFFFFE
	HSTypeConclusion = 0xFFFFFFFF

	HSFlagHSREQ  = 0x00000001
	HSFlagKMREQ  = 0x00000002
	HSFlagCONFIG = 0x00000004

	HSExtTypeHSReq      = 1
	HSExtTypeHSRsp      = 2
	HSExtTypeKMReq      = 3
	HSExtTypeKMRsp      = 4
	HSExtTypeSID        = 5
	HSExtTypeCongestion = 6
	HSExtTypeFilter     = 7
	HSExtTypeGroup      = 8
)
