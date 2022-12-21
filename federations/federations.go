package federations

import (
	"encoding/json"
	"errors"
	"fmt"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/config/arguments"
	"liberty-town/node/contact"
	"liberty-town/node/federations/blockchain-nodes"
	"liberty-town/node/federations/category"
	"liberty-town/node/federations/federation"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/moderator"
	"liberty-town/node/gui"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/store"
	"liberty-town/node/validator"
	"liberty-town/node/validator/validation"
)

//公共网络
var FederationsDict = generics.Map[string, *federation.Federation]{}

//预定义网络
var federationsDescriptive = []map[string]any{

	/**
	├──货物
	│	├──数字
	│	│   ├──加密货币
	│	│   ├──数据
	│	│   ├──信息
	│	│   ├──软件
	│	│   ├──教程
	│	│   ├──网络安全
	│	│   ├──礼品卡
	│	│   └──其他
	│	├──财力
	│	├──钱
	│	├──珠宝
	│	├──金
	│	├──计算机
	│	│   ├──计算机
	│	│   ├──笔记本 电脑
	│	│   ├──产品
	│	│   ├──模块
	│	│   └──电子游戏
	│	├──工具
	│	├──化学药品
	│	├──电子学
	│	├──手机
	│	├──电器
	│	├──汽车
	│	│   ├──车辆
	│	│   ├──越野车
	│	│   ├──卡车
	│	│   └──零件
	│	├──摩托车
	│	│   ├──摩托车
	│	│   └──零件
	│	├──船
	│	│   ├──船
	│	│   └──零件
	│	├──书
	│	├──商
	│	├──艺术
	│	└──服装及配饰
	├──在线服务
	│	├──黑客
	│	├──金融
	│	├──编程服务
	│	├──法律服务
	│	├──数字服务
	│	├──网络服务
	│	│   ├──虚拟主机
	│	│   ├──专用网络
	│	│   ├──代理
	│	│   └──网页设计
	│	├──贸易
	│	└──其他
	├──药
	│	├──类固醇
	│	├──减肥
	│	├──药物
	│	├──补充
	│	├──草药
	│	└──其他
	└──住房
	 	├──房地产
	 	├──公寓
	 	├──房屋互换
	 	├──办公室
	 	├──迷你自存倉
	 	└──共用房间

	*/

	{
		"version":     federation.FEDERATION_VERSION,
		"name":        "Liberty Street",
		"description": "This is the place you always wanted to be.",
		"categories": []*category.Category{
			{0, "Goods", []*category.Category{
				{100000, "Digital", []*category.Category{
					{100010, "Crypto", nil}, {100020, "Data", nil}, {100030, "Information", nil}, {100040, "Software", nil}, {100050, "Tutorials", nil}, {100060, "Security", nil}, {100070, "Gift Cards", nil}, {100080, "Other", nil},
				}},
				{101000, "Financial", nil}, {102000, "Money", nil}, {103000, "Jewellery", nil}, {104000, "Gold", nil},
				{105000, "Computer", []*category.Category{
					{105010, "Computers", nil}, {105020, "Laptops", nil}, {105030, "Gadgets", nil}, {105040, "Parts", nil}, {105050, "Video games", nil},
				}},
				{106000, "Tools", nil}, {107000, "Chemicals", nil}, {108000, "Electronics", nil}, {109000, "Phones", nil}, {11000, "Appliances", nil},
				{111000, "Automotive", []*category.Category{
					{111010, "Vehicles", nil}, {111020, "SUV", nil}, {111030, "Trailers", nil}, {111040, "Parts", nil},
				}},
				{112000, "Motorcycle", []*category.Category{
					{112010, "Motorcycles", nil}, {112020, "Parts", nil},
				}},
				{113000, "Boats", []*category.Category{
					{113010, "Boats", nil}, {113020, "Parts", nil},
				}},
				{114000, "Books", nil}, {115000, "Business", nil}, {116000, "Art", nil},
				{117000, "Clothing & Accessories", nil},
			}},
			{200000, "Services", []*category.Category{
				{201000, "Hacking", nil}, {202000, "Financial", nil}, {203000, "Programming", nil}, {204000, "Legal", nil}, {205000, "Digital", nil},
				{206000, "Web", []*category.Category{
					{206010, "Hosting", nil}, {206020, "VPN", nil}, {206030, "Proxies", nil}, {206040, "Web Design", nil},
				}}, {207000, "Skilled Trade", nil}, {208000, "Other", nil},
			}},
			{300000, "Medicine", []*category.Category{
				{301000, "Steroids", nil}, {302000, "Weight Loss", nil}, {303000, "Medications", nil}, {304000, "Supplements", nil}, {305000, "Herbs", nil}, {306000, "Other", nil},
			}},
			{400000, "Housing", []*category.Category{
				{401000, "Real Estate", nil}, {402000, "Flats", nil}, {403000, "House Swap", nil}, {404000, "Office/commercial", nil}, {405000, "Parking/storage", nil}, {406000, "Shared rooms", nil},
			}},
		},
		"validators": []map[string]any{
			{
				"version":   validator.VALIDATOR_VERSION,
				"contact":   "AAwwLjAuMS10ZXN0LjABACB3c3M6Ly9saWJlcnR5Z2F0ZXdheS5vcmc6MjA4My93c5mx6JsGIOfQG0LLs6cT3JdsrNsLeWJKQlauplM1mXZK2E4B+xEV8GWKQRqQjxgWk1zqdw5OCAd0gmNL+Q4r+PLVsg+1UwE=",
				"ownership": "2bHomwaH+cvKBYn1oWfsUQDQf44osSjbtUSROPfJn+pLC8fcSHoETDoJ6MHBkuADFSEOOzIuoFotGbjrB7p6B1vfrIB0AA==",
			},
			{
				"version":   validator.VALIDATOR_VERSION,
				"contact":   "AAwwLjAuMS10ZXN0LjABAER3czovL2pwNDRvcGpucmZpNGlidzVvdWZuNDRzMjJ5c25lbHRxdjZydzZrZWIyaGM3eGtxMmIzaXEuYjMyLmkycC93c9iPwpwG3J7S7HqtzkyA0lEd4vcT/4S+7YSr4eW5axEYMFBiN0AwaGXrWg8zfmAIWgTFde3Nvtat+qUUN2kUJvgp6cUWSwE=",
				"ownership": "2I/CnAbaqkjeUIS8J39BIBz1zBGVriqAonFu9HyResr+x2Cm0wdTdBSiP2usSEWUQMVSdM/9kZsQzz3mpxFwXbo/1lpIAQ==",
			},
		},
		"seeds": []string{
			//seed 1 public gateway node
			"AAMwLjEBACB3c3M6Ly9saWJlcnR5Z2F0ZXdheS5vcmc6MjA1My93c4Oy6JsGsd1GKDNf7ztSYW+YzLuilIbIHo3h2TpGCzVVmQTBDzwRGC1eYpsnCWoiEfa1C6/wzV+wZE4xWLYFpTgk+XhyZwA=",
			//seed 2 public gateway node mirror via tor
			"AAwwLjAuMS10ZXN0LjABAEd3c3M6Ly96d2Rodm5hdDU0cTY2ZnJzbjJncTZkeHdqdnVmYW54aWNsMjdhcW1ucmRpbXBvaHNvZmtib3phZC5vbmlvbi93c6Xfn5wGMlfnhIDbkBSawBKLc/7kbFrAbehqbGQe9AylsSQWLSBaJg3hn7vXFpLpTzok5VKL6XX0/4O84EOpLBhoIZHb1gA=",
			//seed 3 private tor node
			"AAwwLjAuMS10ZXN0LjABAEd3c3M6Ly9vanRycXdwN2d1ZmRwa2lmbm5pbXNrcmtpbWY3N2ZxNmwzb2t5NTRlNHBqZDR0d2t6ZnZ6cnphZC5vbmlvbi93c/Tfn5wGy1Qr5A9teVz2nyBUflIiqSUCaOHoTbjc4RCmlbxmgLQO93yPZU89q1hxp79dy71QSnix1hlitbmpldQVY9nAnAA=",
			//seed 4 private i2p node
			"AAwwLjAuMS10ZXN0LjABAER3czovL3NjNnA1Z2t2NHlqMmFkNmt3cWpuZ2JybmpvNXl5aXZyYmxtdGdva3Nzdmd0bm1lcW1rYWEuYjMyLmkycC93c6jRvpwGrHKBCVkeiMLhdV5jq3XBTX/wjthWa+g/EYLsyvWxXUMkqTDfIAOGjsXn0hSFVl1PThllnX69pydmf4EnqTrDwQE=",
		},
		"moderators": []map[string]any{
			{
				"version":              moderator.MODERATOR_PANDORA,
				"conditionalPublicKey": "FNWgfjfm4pzDTYZnNY02iOGMuHUMn4dnE0AAwloi67sB",
				"rewardAddresses": []string{
					"PCASH12G8zTxf6YtS2d7QXexbdCS3mra95f4UiiAeKPF167dVtKYqMwyAtLan3r9ehqLMXqfqJxpU6LSBsi68tx63hK6jf6vYhHfL4Uqu9TqZ51JCFKbzkXNkqHsRTMefVeqR6bmDJuVzGVRB",
					"PCASH1RuJ7rKfZFsbQk3sZpwYmHjGo2edwP1sxfyXFy9JRzUKnmJzd5SGf7tXyzvCngRmeCGy9ikdHQ3bPakdp9ZBfCPwiRnCWzkMndUDRPY1XGqmG5QcTb4jH38tS1bwXsgVx375Tzf8hhDY",
					"PCASH1Qc4nCvN438PR1DaZ6nffWLhcEkepEqotd78hRMdN6BjFzUAsiYwbBPqkUDy5TuLKRqXokh78mgfCzvxtvTUStG2vUCHUXu6oqt9LFiDi4JqF4A8Y7ZD6gScQ6Kk6tFfoK7nRPBShspp",
				},
				"fee":        uint64(45000),
				"validation": "APdFc3d3AWat2NDFmwa9N++tQ7vQv49gafZOU3fYVUoV/rHFxjZzHuxVpwq4ex5jgF03iwkTFnKNCh3ypNTlWNn79LgY57NWC2m+PIsEAA==",
				"ownership":  "7NHFmwZUyrc/yHtPouQfA1dEyjZWGG9pF3rRKQQGEOy4ZAVM1i9gsbW36YMKaYZdV9OWi5+g+pITwR/OFOsSKK7VT6fmAQ==",
			},
		},
		"nodes": []string{
			"https://seed.pandoracash.com:2053",
			"https://seed.pandoracash.com:2087",
		},
		"acceptedAssets": []string{
			"PCASH",
		},
		"ownership": "57mJnQZC23+tmKQKCMwBy+uAx7CrUcdKgSL0bSFfAORYqTz/YBHN/BRql+CCjUho7JblzX+FEOIMdmFjwC/m6P9NzeIWAA==",
	},
}

func InitializeFederations() (err error) {

	federations := []*federation.Federation{}

	if config.NETWORK_SELECTED == config.DEV_NET_NETWORK_BYTE {
		federationsDescriptive[0]["validators"] = []map[string]any{
			{
				"version":   validator.VALIDATOR_VERSION,
				"contact":   "AAwwLjAuMS10ZXN0LjABABZ3czovLzEyNy4wLjAuMTo0MDA1L3dz6OXgmwa1CP9D50kjBNAvVp7Tc8Ww4I/vqZcfZk4ZKyaQIuLEbEo3rzAy1kbQs8u1/t+u59FlXu1V9hh6CalJozipCbqkAA==",
				"ownership": "64PhmwalV13CeT91Kh2EuMJDWBL9jAIjPWkCdoP3oe1rJ8trwCJN+aKlLCnanSCnzrs8Y2Zs9s54WenM94jwvCV7CcPoAQ==",
			},
		}
		federationsDescriptive[0]["seeds"] = []string{
			"AAMwLjEBABZ3czovLzEyNy4wLjAuMTo4MDgwL3dzlbjbmwagDIZB18rxFmmC6CD6kj7LnmZNM6jEQIPTNZBQT9nAazo9aVYI4b7ZMHXwrK+VNzcXfpfjOjxqiabKRYOvCRpIAA==",
		}
		federationsDescriptive[0]["moderators"] = []map[string]any{
			{
				"version":              moderator.MODERATOR_PANDORA,
				"conditionalPublicKey": "FNWgfjfm4pzDTYZnNY02iOGMuHUMn4dnE0AAwloi67sB",
				"rewardAddresses": []string{
					"PCASH12G8zTxf6YtS2d7QXexbdCS3mra95f4UiiAeKPF167dVtKYqMwyAtLan3r9ehqLMXqfqJxpU6LSBsi68tx63hK6jf6vYhHfL4Uqu9TqZ51JCFKbzkXNkqHsRTMefVeqR6bmDJuVzGVRB",
					"PCASH1RuJ7rKfZFsbQk3sZpwYmHjGo2edwP1sxfyXFy9JRzUKnmJzd5SGf7tXyzvCngRmeCGy9ikdHQ3bPakdp9ZBfCPwiRnCWzkMndUDRPY1XGqmG5QcTb4jH38tS1bwXsgVx375Tzf8hhDY",
					"PCASH1Qc4nCvN438PR1DaZ6nffWLhcEkepEqotd78hRMdN6BjFzUAsiYwbBPqkUDy5TuLKRqXokh78mgfCzvxtvTUStG2vUCHUXu6oqt9LFiDi4JqF4A8Y7ZD6gScQ6Kk6tFfoK7nRPBShspp",
				},
				"fee":        uint64(45000),
				"validation": "AKxBMoYHE8lp+OXimwbb4BZLZK161nTM3K7aEyVhgJhRy+LICQ0UI1XtZ18yvA+sbq7SVDiB4H4ZpzFCJKO5Mfc61xtqBJxmoAKnP82/AQ==",
				"ownership":  "/eXimwams25zTPJNhf/X9+zPROdLA6g+SHuBvlNm7OFy+jVX7DcWt0sT7+UcGJEs2hRmcg0l0EGQ/hSIYcWM9Gp0aeOyAA==",
			},
		}
		federationsDescriptive[0]["ownership"] = "tObimwZQ3sbHg19g5jhrtW4tJMHZMvvlDE2ZOPeFWVPZ935gfxQUWst8El8HOEePhcpcKeo1f6JWEs5XVG273L2cYROgAA=="
	}

	for _, it := range federationsDescriptive {
		f := &federation.Federation{}
		f.Version = it["version"].(federation.FederationVersion)
		f.Name = it["name"].(string)
		f.Description = it["description"].(string)
		f.Validators = []*validator.Validator{}
		f.Categories = []*category.Category{}

		if it["categories"] != nil {
			f.Categories = it["categories"].([]*category.Category)
		}
		if it["validators"] != nil {
			a := it["validators"].([]map[string]any)
			for _, it2 := range a {
				v := &validator.Validator{}
				v.Version = it2["version"].(validator.ValidatorVersion)
				v.Contact = ContactDeserializeForced(it2["contact"].(string))
				if it2["ownership"] != nil {
					v.Ownership = &ownership.Ownership{}
					OwnershipDeserializedForced(v.Ownership, it2["ownership"].(string), v.GetMessageToSign)
				}
				f.Validators = append(f.Validators, v)
			}
		}
		f.Moderators = []*moderator.Moderator{}
		if it["moderators"] != nil {
			a := it["moderators"].([]map[string]any)
			for _, it2 := range a {
				m := &moderator.Moderator{}
				m.Version = it2["version"].(moderator.ModeratorVersion)
				m.Fee = it2["fee"].(uint64)
				m.ConditionalPublicKey = helpers.DecodeBase64(it2["conditionalPublicKey"].(string))
				if it2["rewardAddresses"] != nil {
					m.RewardAddresses = it2["rewardAddresses"].([]string)
				}
				if it2["validation"] != nil {
					m.Validation = &validation.Validation{}
					if err := ValidationDeserialized(m.Validation, it2["validation"].(string), m.GetMessageForSigningValidator); err != nil {
						return err
					}
				}
				if it2["ownership"] != nil {
					m.Ownership = &ownership.Ownership{}
					OwnershipDeserializedForced(m.Ownership, it2["ownership"].(string), m.GetMessageForSigningOwnership)
				}
				f.Moderators = append(f.Moderators, m)
				if m.Validate() != nil || m.ValidateSignatures() != nil || !f.IsValidationAccepted(m.Validation) {
					return errors.New("invalid moderator")
				}
			}
		}
		f.Seeds = make([]*contact.Contact, 0)
		if it["seeds"] != nil {
			a := it["seeds"].([]string)
			for _, it2 := range a {
				f.Seeds = append(f.Seeds, ContactDeserializeForced(it2))
			}
		}
		f.Nodes = make([]*blockchain_nodes.BlockchainNode, 0)
		if it["nodes"] != nil {
			a := it["nodes"].([]string)
			for _, it2 := range a {
				f.Nodes = append(f.Nodes, &blockchain_nodes.BlockchainNode{it2})
			}
		}
		if it["acceptedAssets"] != nil {
			f.AcceptedAssets = it["acceptedAssets"].([]string)
		}
		if it["ownership"] != nil {
			f.Ownership = &ownership.Ownership{}
			OwnershipDeserializedForced(f.Ownership, it["ownership"].(string), f.GetMessageToSign)
		}
		federations = append(federations, f)
	}

	//调试信息
	if arguments.Arguments["--display-apps"] != "" {
		gui.GUI.Log("")
		for k, f := range federations {
			x, err := json.Marshal(f)
			if err != nil {
				panic(err)
			}
			gui.GUI.Info("FED", k, string(x), "\n")
		}
	}

	for i, it := range federations {

		if err = it.Validate(); err != nil {
			return
		}

		for j, v := range it.Validators {
			if v.Ownership.Address, err = addresses.CreateAddrFromSignature(v.GetMessageToSign(), v.Ownership.Signature); err != nil {
				return fmt.Errorf("invalid federation %d retrieve validator %d address", i, j)
			}
			if !v.Ownership.Verify(v.GetMessageToSign) {
				return fmt.Errorf("invalid federation %d validator %d ownership", i, j)
			}
		}

		if it.Ownership.Address, err = addresses.CreateAddrFromSignature(it.GetMessageToSign(), it.Ownership.Signature); err != nil {
			return fmt.Errorf("invalid federation %d retrieve address", i)
		}
		if it.ValidateSignatures() != nil {
			return fmt.Errorf("invalid federation %d ownership", i)
		}
		FederationsDict.Store(it.Ownership.Address.Encoded, it)

		x := advanced_buffers.NewBufferWriter()
		it.Serialize(x)
		b := x.Bytes()

		fed2 := &federation.Federation{}
		if err := fed2.Deserialize(advanced_buffers.NewBufferReader(b)); err != nil {
			panic(err)
		}

	}

	if store.StoreFederations != nil {
		for _, it := range federations {
			newFed, err := federation_serve.LoadFederation(it.Ownership.Address)
			if err != nil {
				return err
			}
			if newFed != nil {
				FederationsDict.Store(it.Ownership.Address.Encoded, newFed)
			}
		}
	}

	if err = readArgument(); err != nil {
		return
	}

	return nil
}
