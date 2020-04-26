package msg

import (
	"fmt"
	"encoding/json"
	"errors"
)

type Msg struct {
	Tx            interface{} `json:"tx"`
	TxType        string      `json:"tx_type"`
	AccountNumber int64       `json:"account_number"`
	Sequence      int64       `json:"sequence"`
	Fee           string      `json:"fee"`
	Gas           string      `json:"gas"`
	Memo          string      `json:"memo"`
	ChainID       string      `json:"chain_id"`
}
type TransferDetail struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
	From   string `json:"from"`
	To     string `json:"to"`
}
type WithdrawDelegatorRewardDetail struct {
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
}
type DelegateDetail struct {
	Amount    string `json:"amount"`
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
}
type UndelegateDetail struct {
	Amount    string `json:"amount"`
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
}
type RedelegateDetail struct {
	Amount       string `json:"amount"`
	Delegator    string `json:"delegator"`
	ValidatorSrc string `json:"validator_src"`
	ValidatorDst string `json:"validator_dst"`
}
type VoteDetail struct {
	ProposalID int64  `json:"proposal_id"`
	Voter      string `json:"voter"`
	Option     string `json:"option"`
}

func ParseJson(jsonStr string) (msg Msg, err error) {
	err = json.Unmarshal([]byte(jsonStr), &msg)
	if err != nil {
		return
	}
	bz, err := json.Marshal(msg.Tx)
	if err != nil {
		return
	}
	switch msg.TxType {
	case "transfer":
		var tx TransferDetail
		err = json.Unmarshal(bz, &tx)
		msg.Tx = tx
	case "withdraw_delegator_reward":
		var tx WithdrawDelegatorRewardDetail
		err = json.Unmarshal(bz, &tx)
		msg.Tx = tx
	case "delegate":
		var tx DelegateDetail
		err = json.Unmarshal(bz, &tx)
		msg.Tx = tx
	case "undelegate":
		var tx UndelegateDetail
		err = json.Unmarshal(bz, &tx)
		msg.Tx = tx
	case "redelegate":
		var tx RedelegateDetail
		err = json.Unmarshal(bz, &tx)
		msg.Tx = tx
	case "vote":
		var tx VoteDetail
		err = json.Unmarshal(bz, &tx)
		if tx.Option != "Yes" && tx.Option != "No" && tx.Option != "Abstain" && tx.Option != "NoWithVeto" {
			err = errors.New("Invalid Option: "+tx.Option)
		}
		msg.Tx = tx
	case "raw":
		var raw []string
		err = json.Unmarshal(bz, &raw)
		if err != nil && len(raw) != 2 {
			err = errors.New("Invalid tx for 'raw': it is not a string slice")
		}
		msg.Tx = raw
	default:
		err = errors.New("Invalid tx_type: "+msg.TxType)
	}
	return
}

//type StdSignDoc struct {
//	AccountNumber uint64            `json:"account_number" yaml:"account_number"`
//	ChainID       string            `json:"chain_id" yaml:"chain_id"`
//	Fee           json.RawMessage   `json:"fee" yaml:"fee"`
//	Memo          string            `json:"memo" yaml:"memo"`
//	Msgs          []json.RawMessage `json:"msgs" yaml:"msgs"`
//	Sequence      uint64            `json:"sequence" yaml:"sequence"`
//}
func (msg Msg) GetSignBytes() []byte {
	MsgSend := `[{"type":"bankx/MsgSend","value":{"from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"%s"}],"unlock_time":"0"}}]`
	MsgWithdrawDelegationReward := `[{"type":"cosmos-sdk/MsgWithdrawDelegationReward","value":{"delegator_address":"%s","validator_address":"%s"}}]`
	MsgDelegate := `[{"type":"cosmos-sdk/MsgDelegate","value":{"delegator_address":"%s","validator_address":"%s","amount":{"denom":"cet","amount":"%s"}}}]`
	MsgUndelegate := `[{"type":"cosmos-sdk/MsgUndelegate","value":{"delegator_address":"%s","validator_address":"%s","amount":{"denom":"cet","amount":"%s"}}}]`
	MsgBeginRedelegate := `[{"type":"cosmos-sdk/MsgBeginRedelegate","value":{"delegator_address":"%s","validator_src_address":"%s","validator_dst_address":"%s","amount":{"denom":"cet","amount":"%s"}}}]`
	MsgVote := `[{"type":"cosmos-sdk/MsgVote","value":{"proposal_id":"%d","voter":"%s","option":"%s"}}]`

	var msgStr string
	switch tx := msg.Tx.(type) {
	case TransferDetail:
		msgStr = fmt.Sprintf(MsgSend, tx.From, tx.To, tx.Denom, tx.Amount)
	case WithdrawDelegatorRewardDetail:
		msgStr = fmt.Sprintf(MsgWithdrawDelegationReward, tx.Delegator, tx.Validator)
	case DelegateDetail:
		msgStr = fmt.Sprintf(MsgDelegate, tx.Delegator, tx.Validator, tx.Amount)
	case UndelegateDetail:
		msgStr = fmt.Sprintf(MsgUndelegate, tx.Delegator, tx.Validator, tx.Amount)
	case RedelegateDetail:
		msgStr = fmt.Sprintf(MsgBeginRedelegate, tx.Delegator, tx.ValidatorSrc, tx.ValidatorDst, tx.Amount)
	case VoteDetail:
		msgStr = fmt.Sprintf(MsgVote, tx.ProposalID, tx.Voter, tx.Option)
	case []string:
		msgStr = tx[1]
	}
	doc := `{"account_number":%d,"chain_id":%d,"fee":{"amount":[{"denom":"cet","amount":"%s"}],"gas":"%s"},"memo":"%s","msg":%s,"sequence":%d}`
	res := fmt.Sprintf(doc, msg.AccountNumber, msg.ChainID, msg.Fee, msg.Gas, msg.Memo, msgStr, msg.Sequence)
	return []byte(res)
}

func (msg Msg) GetSigner() string {
	switch tx := msg.Tx.(type) {
	case TransferDetail:
		return tx.From
	case WithdrawDelegatorRewardDetail:
		return tx.Delegator
	case DelegateDetail:
		return tx.Delegator
	case UndelegateDetail:
		return tx.Delegator
	case RedelegateDetail:
		return tx.Delegator
	case VoteDetail:
		return tx.Voter
	case []string:
		return tx[1]
	default:
		return ""
	}
}

