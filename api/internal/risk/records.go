package risk

import (
	"context"
	"encoding/json"
	"time"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type Store interface {
	Profile(context.Context, string, string) (SavedProfile, error)
	SaveProfile(context.Context, string, string, Profile) (SavedProfile, error)
	RecordCheck(context.Context, string, string, Request, *enginev1.CheckResponse) error
	Checks(context.Context, string, string, int) ([]CheckRecord, error)
}

type CheckRecord struct {
	ID            string    `json:"id"`
	Source        string    `json:"source"`
	OrderID       *string   `json:"order_id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	EntryPrice    string    `json:"entry_price"`
	Allowed       bool      `json:"allowed"`
	Verdict       string    `json:"verdict"`
	FailedRules   []string  `json:"failed_rules"`
	CautionRules  []string  `json:"caution_rules"`
	ProfileSource string    `json:"profile_source"`
	CreatedAt     time.Time `json:"created_at"`
}

func ProfileSource(body Request) string {
	if body.RiskProfile != nil {
		return "preview"
	}
	return "stored"
}

func MarshalCheck(result *enginev1.CheckResponse) ([]byte, error) {
	return (protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}).Marshal(result)
}

func checkResponse(result *enginev1.CheckResponse, source string) (map[string]json.RawMessage, error) {
	encoded, err := MarshalCheck(result)
	if err != nil {
		return nil, err
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &value); err != nil {
		return nil, err
	}
	value["profile_source"], err = json.Marshal(source)
	return value, err
}
