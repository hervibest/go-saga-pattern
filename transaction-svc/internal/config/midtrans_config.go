package config

import (
	"go-saga-pattern/commoner/utils"

	"github.com/midtrans/midtrans-go"

	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransClient struct {
	Snap    *snap.Client
	CoreApi *coreapi.Client
}

func NewMidtransClient() *MidtransClient {
	var midtransKey string
	if IsLocal() || IsDevelopment() {
		midtransKey = utils.GetEnv("MIDTRANS_DEV_SERVER_KEY")
	} else if IsProduction() {
		midtransKey = utils.GetEnv("MIDTRANS_PROD_SERVER_KEY")
	}

	snap := &snap.Client{}
	snap.New(midtransKey, midtrans.Sandbox)

	coreApi := &coreapi.Client{}
	coreApi.New(midtransKey, midtrans.Sandbox)

	return &MidtransClient{
		Snap:    snap,
		CoreApi: coreApi,
	}
}
