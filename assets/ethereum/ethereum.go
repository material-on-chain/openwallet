package ethereum

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/console"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/timer"
	"github.com/bndr/gotabulate"
)

type WalletManager struct{}

const (
	TRANS_AMOUNT_UNIT_LIST = `
	1: wei
	2: Kwei
	3: Mwei
	4: GWei
	5: microether
	6: milliether
	7: ether
	`
	TRANS_AMOUNT_UNIT_WEI          = 1
	TRANS_AMOUNT_UNIT_K_WEI        = 2
	TRANS_AMOUNT_UNIT_M_WEI        = 3
	TRANS_AMOUNT_UNIT_G_WEI        = 4
	TRANS_AMOUNT_UNIT_MICRO_ETHER  = 5
	TRANS_AMOUNT_UNIT_MILLIE_ETHER = 6
	TRNAS_AMOUNT_UNIT_ETHER        = 7
)

func toHexBigIntForEtherTrans(value string, base int, unit int64) (*big.Int, error) {
	amount, err := convertToBigInt(value, base)
	if err != nil {
		openwLogger.Log.Errorf("format transaction value failed, err = %v", err)
		return big.NewInt(0), err
	}

	switch unit {
	case TRANS_AMOUNT_UNIT_WEI:
	case TRANS_AMOUNT_UNIT_K_WEI:
		amount.Mul(amount, big.NewInt(1000))
	case TRANS_AMOUNT_UNIT_M_WEI:
		amount.Mul(amount, big.NewInt(1000*1000))
	case TRANS_AMOUNT_UNIT_G_WEI:
		amount.Mul(amount, big.NewInt(1000*1000*1000))
	case TRANS_AMOUNT_UNIT_MICRO_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000))
	case TRANS_AMOUNT_UNIT_MILLIE_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000*1000))
	case TRNAS_AMOUNT_UNIT_ETHER:
		amount.Mul(amount, big.NewInt(1000*1000*1000*1000*1000*1000))
	default:
		return big.NewInt(0), errors.New("wrong unit inputed")
	}

	return amount, nil
}

//初始化配置流程
func (this *WalletManager) InitConfigFlow() error {
	file := filepath.Join(configFilePath, configFileName)
	fmt.Printf("You can run 'vim %s' to edit wallet's config.\n", file)
	return nil
}

//查看配置信息
func (this *WalletManager) ShowConfig() error {
	return printConfig()
}

//创建钱包流程
func (this *WalletManager) CreateWalletFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	// 等待用户输入钱包名字
	name, err := console.InputText("Enter wallet's name: ", true)

	// 等待用户输入密码
	password, err := console.InputPassword(true, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err = %v", err)
		return err
	}

	_, keyFile, err := CreateNewWallet(name, password)
	if err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("Wallet create successfully, key path: %s\n", keyFile)

	return nil
}

/*
type ERC20Token struct {
	Address  string `json:"address" storm:"id"`
	Symbol   string `json:"symbol" storm:"index"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	Valid    int    `json:"valid"`
	balance  *big.Int
}
*/

func printTokenAvailable(list []ERC20Token) {
	tableInfo := make([][]interface{}, 0)

	for i, w := range list {
		tableInfo = append(tableInfo, []interface{}{
			i, w.Symbol, w.Name, w.Address, w.Name,
		})
	}
	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "Symbol", "Name", "Contract Address", "Decimals"})

	//打印信息
	fmt.Println(t.Render("simple"))
}

func printTokenWalletList(list []*Wallet) {
	tableInfo := make([][]interface{}, 0)

	for i, w := range list {

		tableInfo = append(tableInfo, []interface{}{
			i, w.WalletID, w.Alias, w.erc20Token.Symbol, w.erc20Token.balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "Symbol", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))
}

//打印钱包列表
func printWalletList(list []*Wallet) {

	tableInfo := make([][]interface{}, 0)

	for i, w := range list {

		tableInfo = append(tableInfo, []interface{}{
			i, w.WalletID, w.Alias, w.balance,
		})
	}

	t := gotabulate.Create(tableInfo)
	// Set Headers
	t.SetHeaders([]string{"No.", "ID", "Name", "Balance"})

	//打印信息
	fmt.Println(t.Render("simple"))

}

//创建地址流程
func (this *WalletManager) CreateAddressFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//查询所有钱包信息
	wallets, err := GetWalletList()
	if err != nil {
		fmt.Printf("The node did not create any wallet!\n")
		return err
	}

	//打印钱包
	printWalletList(wallets)

	fmt.Printf("[Please select a wallet account to create address] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(wallets) {
		return errors.New("Input number is out of index! ")
	}

	account := wallets[num]

	// 输入地址数量
	count, err := console.InputNumber("Enter the number of addresses you want: ", false)
	if err != nil {
		return err
	}

	if count > maxAddresNum {
		return errors.New(fmt.Sprintf("The number of addresses can not exceed %d\n", maxAddresNum))
	}

	//输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err = %v", err)
		return err
	}

	err = UnlockWallet(account, password)
	if err != nil {
		openwLogger.Log.Errorf("unlock wallet [%v] failed, err = %v", account.WalletID, err)
		return err
	}

	log.Printf("Start batch creation\n")
	log.Printf("-------------------------------------------------\n")

	err = CreateBatchAddress(account.WalletID, password, count)
	if err != nil {
		return err
	}

	log.Printf("-------------------------------------------------\n")
	log.Printf("All addresses have created, file path:%s\n", EthereumKeyPath)

	return nil
}

func (this *WalletManager) ERC20TokenSummaryFollow() error {
	endRunning := make(chan bool, 1)
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(sumAddress) == 0 {
		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", Symbol))
	}

	ercTokens, err := GetERC20TokenList()
	if err != nil {
		openwLogger.Log.Errorf("find tokens failed, err = %v", err)
		return err
	}

	if len(ercTokens) == 0 {
		openwLogger.Log.Errorf("no token available, config the tokens first.")
		return err
	}

	printTokenAvailable(ercTokens)

	//选择token
	tokenId, err := console.InputNumber("Enter Token No. : ", true)
	if err != nil {
		return err
	}

	if int(tokenId) >= len(ercTokens) {
		return errors.New("Input Token No. is out of index! ")
	}

	token := ercTokens[tokenId]
	fmt.Println("token[", token.Symbol, "] is chosen. ")

	wallets, err := ERC20GetWalletList(&token)
	if err != nil {
		return err
	}

	//打印钱包列表
	printTokenWalletList(wallets)

	fmt.Printf("[Please select the wallet to summary, and enter the numbers split by ','." +
		" For example: 0,1,2,3] \n")

	// 等待用户输入钱包名字
	nums, err := console.InputText("Enter the No. group: ", true)
	if err != nil {
		return err
	}

	//分隔数组
	array := strings.Split(nums, ",")

	for _, numIput := range array {
		if common.IsNumberString(numIput) {
			numInt := common.NewString(numIput).Int()
			if numInt < len(wallets) {
				w := wallets[numInt]
				fmt.Printf("Register summary wallet [%s]-[%s]\n", w.Alias, w.WalletID)

				password, err := console.InputPassword(false, 8)
				if err != nil {
					openwLogger.Log.Errorf("input wallet password failed, err=%v", err)
					return err
				}

				err = UnlockWallet(w, password)
				if err != nil {
					openwLogger.Log.Errorf("unlock wallet [%v] failed, err = %v", w.WalletID, err)
					return err
				}
				w.Password = password
				AddWalletInSummary(w.WalletID, w)
			} else {
				return errors.New("The input No. out of index! ")
			}
		} else {
			return errors.New("The input No. is not numeric! ")
		}
	}

	if len(walletsInSum) == 0 {
		return errors.New("Not summary wallets to register! ")
	}

	fmt.Printf("The timer for summary has started. Execute by every %v seconds.\n", cycleSeconds.Seconds())

	//启动钱包汇总程序
	//sumTimer := timer.NewTask(cycleSeconds, ERC20SummaryWallets)
	//sumTimer.Start()
	go ERC20SummaryWallets()

	<-endRunning
	return nil
}

//汇总钱包流程
func (this *WalletManager) SummaryFollow() error {
	endRunning := make(chan bool, 1)
	err := loadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(sumAddress) == 0 {
		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", Symbol))
	}

	wallets, err := GetWalletList()
	if err != nil {
		return err
	}

	//打印钱包列表
	printWalletList(wallets)

	fmt.Printf("[Please select the wallet to summary, and enter the numbers split by ','." +
		" For example: 0,1,2,3] \n")

	// 等待用户输入钱包名字
	nums, err := console.InputText("Enter the No. group: ", true)
	if err != nil {
		return err
	}

	//分隔数组
	array := strings.Split(nums, ",")

	for _, numIput := range array {
		if common.IsNumberString(numIput) {
			numInt := common.NewString(numIput).Int()
			if numInt < len(wallets) {
				w := wallets[numInt]
				fmt.Printf("Register summary wallet [%s]-[%s]\n", w.Alias, w.WalletID)

				password, err := console.InputPassword(false, 8)
				if err != nil {
					openwLogger.Log.Errorf("input wallet password failed, err=%v", err)
					return err
				}

				err = UnlockWallet(w, password)
				if err != nil {
					openwLogger.Log.Errorf("unlock wallet [%v] failed, err = %v", w.WalletID, err)
					return err
				}
				w.Password = password
				AddWalletInSummary(w.WalletID, w)
			} else {
				return errors.New("The input No. out of index! ")
			}

		} else {
			return errors.New("The input No. is not numeric! ")
		}
	}

	if len(walletsInSum) == 0 {
		return errors.New("Not summary wallets to register! ")
	}

	fmt.Printf("The timer for summary has started. Execute by every %v seconds.\n", cycleSeconds.Seconds())

	//启动钱包汇总程序
	sumTimer := timer.NewTask(cycleSeconds, SummaryWallets)
	sumTimer.Start()
	//go SummaryWallets()

	<-endRunning
	return nil
}

//查看钱包列表，显示信息
func (this *WalletManager) GetWalletList() error {
	err := loadConfig()
	if err != nil {
		return err
	}

	//判断汇总地址是否存在
	if len(sumAddress) == 0 {
		return errors.New(fmt.Sprintf("Summary address is not set. Please set it in './conf/%s.ini' \n", Symbol))
	}

	wallets, err := GetWalletList()
	if err != nil {
		return err
	}

	//打印钱包列表
	printWalletList(wallets)
	return nil
}

func (this *WalletManager) ConfigERC20Token() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	ercTokens, err := GetERC20TokenList()
	if err != nil {
		openwLogger.Log.Errorf("find tokens failed, err = %v", err)
		return err
	}

	if len(ercTokens) == 0 {
		openwLogger.Log.Errorf("no token available, config the tokens first.")
	}

	printTokenAvailable(ercTokens)

	tokenName, err := console.InputText("Enter Token Name. : ", true)
	if err != nil {
		return err
	}

	tokenSymbol, err := console.InputText("Enter Token Symbol. : ", true)
	if err != nil {
		return err
	}

	tokenAddress, err := console.InputText("Enter Token Address. : ", true)
	if err != nil {
		return err
	}

	tokenDecimal, err := console.InputNumber("Enter Token Decimals. :", false)
	if err != nil {
		return err
	}

	tosave, err := console.InputText("Save Token Config [Y/N]. :", true)
	if err != nil {
		return err
	}
	tosave = strings.ToLower(tosave)
	if tosave != "y" {
		fmt.Println("give up token config. ")
		return nil
	}

	tokenConfig := &ERC20Token{
		Name:     tokenName,
		Symbol:   tokenSymbol,
		Address:  tokenAddress,
		Decimals: int(tokenDecimal),
	}

	err = SaveERC20TokenConfig(tokenConfig)
	if err != nil {
		openwLogger.Log.Errorf("save token config failed, err = %v", err)
		return err
	}

	ercTokens, err = GetERC20TokenList()
	if err != nil {
		openwLogger.Log.Errorf("find tokens failed, err = %v", err)
		return err
	}

	if len(ercTokens) == 0 {
		openwLogger.Log.Errorf("no token available, config the tokens first.")
		return err
	}

	printTokenAvailable(ercTokens)

	return nil
}

func (this *WalletManager) ERC20TokenTransferFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	ercTokens, err := GetERC20TokenList()
	if err != nil {
		openwLogger.Log.Errorf("find tokens failed, err = %v", err)
		return err
	}

	if len(ercTokens) == 0 {
		openwLogger.Log.Errorf("no token available, config the tokens first.")
		return err
	}

	printTokenAvailable(ercTokens)

	//选择钱包
	tokenId, err := console.InputNumber("Enter Token No. : ", true)
	if err != nil {
		return err
	}

	if int(tokenId) >= len(ercTokens) {
		return errors.New("Input Token No. is out of index! ")
	}

	token := ercTokens[tokenId]
	fmt.Println("token[", token.Symbol, "] is chosen. ")

	list, err := ERC20GetWalletList(&token)
	if err != nil {
		return err
	}

	printTokenWalletList(list)

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(list) {
		return errors.New("Input number is out of index! ")
	}

	wallet := list[num]
	fmt.Println("wallet[", wallet.Alias, "] is chosen. ")

	// 等待用户输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err=%v", err)
		return err
	}

	// 等待用户输入发送数量
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	value, err := console.InputRealNumber("Enter amount to send : ", true)
	if err != nil {
		return err
	}

	amount, err := convertToBigInt(value, 10)
	if err != nil {
		openwLogger.Log.Errorf("convert to big int failed, err = %v", err)
		return err
	}

	fmt.Println("amount input:", amount.String())

	if wallet.erc20Token.balance.Cmp(amount) < 0 {
		return errors.New("Input amount is greater than balance! ")
	}

	//建立交易单
	txID, err := ERC20SendTransaction(wallet,
		receiver, amount, password, true)
	if err != nil {
		return err
	}

	fmt.Printf("Send transaction successfully, TXID：%s\n", txID)

	return nil
}

//发送交易
func (this *WalletManager) TransferFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	list, err := GetWalletList()
	if err != nil {
		return err
	}

	//打印钱包列表
	printWalletList(list)

	fmt.Printf("[Please select a wallet to send transaction] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(list) {
		return errors.New("Input number is out of index! ")
	}

	wallet := list[num]
	fmt.Println("wallet[", wallet.Alias, "] is chosen. ")

	// 等待用户输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err=%v", err)
		return err
	}

	// 等待用户输入发送数量
	receiver, err := console.InputText("Enter receiver address: ", true)
	if err != nil {
		return err
	}

	fmt.Println("receiver: ", receiver)

	fmt.Println("Choose the unit for the transaction:")
	fmt.Println(TRANS_AMOUNT_UNIT_LIST)
	unit, err := console.InputNumber("Index of the unit: ", true)
	if err != nil {
		return err
	}

	amount, err := console.InputRealNumber("Enter amount to send : ", true)
	if err != nil {
		return err
	}

	amountInt, err := toHexBigIntForEtherTrans(amount, 10, int64(unit))
	if err != nil {
		openwLogger.Log.Errorf("wrong amount inputed. ")
		return err
	}

	fmt.Println("amount input:", amountInt.String())

	if wallet.balance.Cmp(amountInt) < 0 {
		return errors.New("Input amount is greater than balance! ")
	}

	//建立交易单
	txID, err := SendTransaction(wallet,
		receiver, amountInt, password, true)
	if err != nil {
		return err
	}

	fmt.Printf("Send transaction successfully, TXID：%s\n", txID)

	return nil
}

//备份钱包流程
func (this *WalletManager) BackupWalletFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	wallets, err := GetWalletList()
	if err != nil {
		openwLogger.Log.Errorf("get wallet list failed, err = ", err)
		return err
	}

	//打印钱包列表
	printWalletList(wallets)

	fmt.Printf("[Please select a wallet to backup] \n")

	//选择钱包
	num, err := console.InputNumber("Enter wallet No. : ", true)
	if err != nil {
		return err
	}

	if int(num) >= len(wallets) {
		return errors.New("Input number is out of index! ")
	}

	wallet := wallets[num]

	// 等待用户输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err = %v", err)
		return err
	}

	backupPath, err := BackupWalletToDefaultPath(wallet, password)
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Wallet backup file path: %s\n", backupPath)

	return nil
}

//恢复钱包
func (this *WalletManager) RestoreWalletFlow() error {
	//先加载是否有配置文件
	err := loadConfig()
	if err != nil {
		return err
	}

	//输入恢复文件路径
	keyPath, err := console.InputText("Enter backup key file path: ", true)
	if err != nil {
		return err
	}

	// 等待用户输入密码
	password, err := console.InputPassword(false, 8)
	if err != nil {
		openwLogger.Log.Errorf("input password failed, err = %v", err)
		return err
	}

	fmt.Printf("Wallet restoring, please wait a moment...\n")
	err = RestoreWallet(keyPath, password)
	if err != nil {
		return err
	}

	//输出备份导出目录
	fmt.Printf("Restore wallet successfully.\n")

	return nil
}