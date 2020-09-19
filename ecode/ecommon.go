package ecode

// All common ecode
var (
	OK = add(0) // 正确

	AccessKeyErr       = add(-2) // Access Key错误
	SignCheckErr       = add(-3) // API校验密匙错误
	MethodNoPermission = add(-4) // 调用方对该Method没有权限
	ParamErr           = add(-5) //参数错误
	ContextTypeErr     = add(-6) //ContextType不对

	UsernameNotExist         = add(-10) //用户名不存在
	EmailNotExist            = add(-11) //邮箱不存在
	MobileNotExist           = add(-12) //手机号不存在
	PasswordErr              = add(-13) //密码错误
	TwoFAErr                 = add(-14) //二次验证错误
	AccountNotExist          = add(-15) //账户不存在
	UsernameInvalid          = add(-16) //用户名无效
	EmailInvalid             = add(-17) //邮箱无效
	MobileInvalid            = add(-18) //手机号无效
	UsernameHasExist         = add(-19) //用户名已存在
	EmailHasExist            = add(-20) //邮箱已存在
	MobileHasExist           = add(-21) //手机号已存在
	ActivationCodeNotExist   = add(-22) //激活码不存在
	ActivationCodeInvalid    = add(-23) //激活码已失效
	TwoFAInvalid             = add(-24) //二次验证码已失效
	UserUnActive             = add(-25) //用户未激活
	UserOrAccountHasFreezed  = add(-26) //用户或账户被冻结
	RoleNotExist             = add(-27) //角色不存在
	PermNotExist             = add(-28) //权限不存在
	ActivationCodeNotifyFail = add(-29) //激活码通知发送失败
	NeedFACaptha             = add(-30) //需要二次验证码
	UserLoginLock            = add(-31) //登录密码错误重试太多已被锁定
	PasswordNoSet            = add(-32) //密码未设置
	TwoFARepeat              = add(-33) //已使用过的二次验证码
	NoRelation               = add(-34) // 试图操作不相关用户

	NotExistOrNoRelation = add(-35) // 要操作的数据不存在或非用户所有

	CoinWalletNotExist = add(-50) // coin wallet 不存在
	CoinNotExist       = add(-51) // coin不存在

	NoLogin                 = add(-101) //账号未登录
	UserDisabled            = add(-102) //账号被封停
	CaptchaErr              = add(-105) //验证码错误
	UserInactive            = add(-106) //账号未激活
	MobileNoVerfiy          = add(-110) //未绑定手机
	CsrfNotMatchErr         = add(-111) //csrf 校验失败
	ServiceUpdate           = add(-112) //系统升级中
	UserIDCheckInvalid      = add(-113) //账号尚未实名认证
	UserIDCheckInvalidPhone = add(-114) //请先绑定手机
	NeedOtp                 = add(-115) //需要二次验证码

	RecordNotExist    = add(-302) //记录不存在
	RecordHasExist    = add(-303) //记录已经存在
	NotModified       = add(-304) //木有改动
	TemporaryRedirect = add(-307) //撞车跳转
	RequestErr        = add(-400) //请求错误
	Unauthorized      = add(-401) //未认证
	AccessDenied      = add(-403) //访问权限不足
	NothingFound      = add(-404) //啥都木有
	MethodNotAllowed  = add(-405) //不支持该方法
	Conflict          = add(-409) //冲突
	Canceled          = add(-498) //客户端取消请求

	ServerErr          = add(-500) //服务器错误
	ServiceUnavailable = add(-503) //过载保护,服务暂不可用
	Deadline           = add(-504) //服务调用超时
	LimitExceed        = add(-509) //超出限制

	AssetInsufficient     = add(-600) //资产余额不足
	FeeInsufficient       = add(-601) //手续费不足
	AddressNotInWhitelist = add(-602) //地址不在白名单里
	OverLimit             = add(-603) //超出限额
	AccountAssetError     = add(-604) //资产为负，出错了

	WalletCoinNotExist     = add(-700) //钱包未添加此代币
	ForbidWithdraw         = add(-701) //禁止提现
	ForbidDeposit          = add(-702) //禁止充值
	CoinNotActive          = add(-703) //币种下架
	AddressNotActive       = add(-704) //地址失效
	ForbidNewAddress       = add(-705) //禁止生成新地址
	AddressExist           = add(-706) //地址已存在
	AddressInvalid         = add(-707) //地址无效
	FeeZero                = add(-708) //手续费不能为0
	WalletNotExist         = add(-709) //钱包不存在
	AddressNotExist        = add(-710) //地址不存在
	WalletAssetNotExist    = add(-711) //钱包资产不存在
	WalletTypeErr          = add(-712) //钱包类型不对
	AddressNotInner        = add(-713) //不是钱包内部地址
	TxNotExist             = add(-714) //Tx 不存在
	WalletCoinAlreadyExist = add(-715) // coin wallet已经存在
	WalletCoinHasAddress   = add(-716) // coin wallet已有地址，不能删除

	WalletHasWalletCoin = add(-717) // 钱包已有Wallet Coin，不能删除
	AddressCreateFail   = add(-718) // 地址创建失败

)
