package main

import (
	"github.com/cloudfoundry/jibber_jabber"
	"github.com/qor/i18n"
)

var LC string

var I18n = i18n.New()

func T(s string) string {
	return string(I18n.T(LC, s))
}

func add(key, en, cn string) {
	I18n.AddTranslation(&i18n.Translation{
		Key:    key,
		Locale: "en-US",
		Value:  en,
	})
	I18n.AddTranslation(&i18n.Translation{
		Key:    key,
		Locale: "zh-CN",
		Value:  cn,
	})
}

func init() {
	LC, _ = jibber_jabber.DetectIETF() //zh-CN en-US
	LC = "en-US"
	add("appName", "CoinEx Chain Cold Wallet", "CoinEx Chain 冷钱包")
	add("ok", "OK", "确认")
	add("cancel", "Cancel", "取消")
	add("delete", "Delete", "删除")
	add("exit", "Exit", "退出程序")
	add("open", "Open Keybase", "打开私钥数据库")
	add("create&open", "Create and Open a Keybase", "创建并打开私钥数据库")
	add("file", "File", "文件")
	add("help", "Help", "帮助")
	add("scanQRCode", "Scan QRCode", "扫描二维码")
	add("hideWin", "Hide Window", "隐藏窗口")
	add("about", "About", "关于")
	add("success", "Success", "成功")
	add("error!", "Error!", "错误！")
	add("alreadyRun", "An instance of this program is already running", "本程序的一个实例已经在运行了")
	add("notOpen", "Keybase is not opend", "尚未打开私钥数据库")
	add("kbNotOpen", "Keybase has not been opened. Please open or create one.", "私钥数据库尚未打开，请打开一个私钥数据库文件，或者创建一个。")
	add("aboutTitle", "Cold Wallet for CoinEx Chain", "CoinEx Chain 冷钱包")
	add("aboutContent", "Create, manage accounts and sign transactions with accounts' private keys",
		"创建、管理账户，以及使用账户的私钥来签署交易")
	add("selectOpen", "Please Select a File containing the Keybase", "请选择一个包含私钥数据库的文件")
	add("selectSave", "Please Select a File to contain the Keybase", "请选择一个文件用于保存私钥数据库")
	add("ca", "Create Account", "创建账户")
	add("la", "List Account", "账户列表")
	add("sign", "Sign", "签名")
	add("changePassphrase", "Change Passphrase", "更改口令")
	add("showMnemonic", "Show Mnemonic", "显示助记词")
	add("showQRCode", "Show QR Code for Signed Result", "将签名结果显示为二维码")
	add("showAddrQRCode", "Show QR Code of Address", "显示地址的二维码")
	add("copyAddr", "Copy Address", "复制地址")
	add("invalid_prefix", "Invalid Prefix!", "非法前缀！")
	add("invalid_char", "Invalid Character: %s\n", "非法字符：%s\n")
	add("invalid_suffix", "Invalid Suffix!", "非法后缀！")
	add("warn", "Warning", "警告")
	add("long_run_time", "You specified %d characters totally. It would take very long time to compute!", "您总共指定了%d个字符，这需要很长的时间才能生成一个靓号！")
	add("estimate_progress", "%d times have been tried, estimated progress: %.2f%%\r\n", "已经进行了%d次尝试，估计的完成度为%.2f%%\r\n")
	add("prefix", "Prefix", "前缀")
	add("suffix", "Suffix", "后缀")
	add("encryptPassphrase", "Passphrase for Encryption", "加密口令")
	add("failOpenWebcam", "Failed to open the Webcam", "摄像头打开失败")
	add("retypeEncryptPassphrase", "Retype Passphrase for Encryption", "再次确认加密口令")
	add("enterEncryptPassphrase", "Enter the Passphrase for Encryption", "请输入加密口令")
	add("changeEncryptPassphrase", "Change the Passphrase for Encryption", "更新加密口令")
	add("enterOldEncryptPassphrase", "Enter the Old Passphrase for Encryption", "请输入旧的加密口令")
	add("belowNewEncryptPassphrase", "Enter the New Passphrase Below", "在下方输入新的口令")
	add("memo", "Memo", "备忘")
	add("progress", "Progress", "进展")
	add("caline1", "Please enter the prefix and suffix of your desired address below.",
		"请在下方输入您所期待的地址的前缀和后缀。")
	add("caline2", "Prefix is the characters coming immediately after \"coinex1\".",
		"前缀是指紧接着\"coinex1\"出现的若干字符。")
	add("caline3", "Suffix is the last few characters at the ending of an address.",
		"后缀是指地址末尾的若干字符。")
	add("caline4", "Valid characters in a bech32 address are '023456789acdefghjklmnpqrstuvwxyz'.",
		"请注意在bech32格式的地址中，只有这些字符是合法的：'023456789acdefghjklmnpqrstuvwxyz'")
	add("caline5", "Please note 'b', 'i', 'o' and '1' are not valid after \"coinex1\".",
		"也就是说，这几个字符不能出现在\"coinex1\"之后：b, i, o, 1")
	add("caline6", "The passphrase is used to encrypt the private key stored on disk.",
		"加密口令将被用来加密存储在磁盘上的私钥。")
	add("caline7", "Memo is some information to remind yourself what's the usage of this account.",
		"备忘一栏用来填写一些信息，用来提醒你自己这个账户的用途是什么")
	add("origMsg", "Original Message", "原始消息")
	add("readableMsg", "Readable Message", "可读的消息")
	add("mismatchPassphrase", "The two passphrases are mismatched", "输入的两个口令不一致")
	add("emptyMemo", "The Memo can not be empty", "输入的备忘信息不可以为空")
	add("sureToExit", "Are you sure to exit this program?", "您确认要退出本程序吗？")
	add("exit?", "Exit?", "退出？")
	add("noAccYet", "The Keybase has no accounts yet.", "私钥数据库中尚未创建任何账户。")
	add("noSelAcc", "No account was selected in the list.", "您尚未选中列表中的任一账户。")
	add("mnemonicOf", "The mnemonics of %s", "%的助记词")
	add("qrCodeOfAddr", "The QRCode of \"%s\"", "\"%s\"的二维码")
	add("qrCodeOfAddrBelow", "Below is the QRCode of \"%s\"", "下面是\"%s\"的二维码")
	add("successCA", "Success in creating an account: ", "账户创建成功：")
	add("successCopy", "Success in copying address", "账户地址已成功拷贝")
	add("successSign", "Success in signing", "签名成功")
	add("notHaveAcc", "This Keybase does not have the required account: ", "此私钥数据库并未保存签名所需要的账户：")
	add("seeSignQRBelow",
	"The signed result is ready. You can scan it from the below QRCode",
	"签名结果已生成，您可以从下面到二维码中扫描得到它：")
}

