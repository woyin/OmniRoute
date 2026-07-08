package registry

// RegisterRegionalProviders adds regional and enterprise providers.
func RegisterRegionalProviders() {
	Register(&RegistryEntry{
		ID:         "baidu",
		Name:       "Baidu Qianfan",
		BaseURL:    "https://aip.baidubce.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "tencent",
		Name:       "Tencent Cloud",
		BaseURL:    "https://hunyuan.tencentcloudapi.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "iflytek",
		Name:       "iFlytek Spark",
		BaseURL:    "https://spark-api.xf-yun.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "sparkdesk",
		Name:       "SparkDesk",
		BaseURL:    "https://spark-api.xf-yun.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "stepfun",
		Name:       "StepFun",
		BaseURL:    "https://api.stepfun.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "sensenova",
		Name:       "SenseNova",
		BaseURL:    "https://api.sensenova.cn",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "qiniu",
		Name:       "Qiniu AI",
		BaseURL:    "https://api.qiniu.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "digitalocean",
		Name:       "DigitalOcean AI",
		BaseURL:    "https://inference.do.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "oci",
		Name:       "Oracle OCI AI",
		BaseURL:    "https://inference.generativeai.ocp.oraclecloud.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "watsonx",
		Name:       "IBM watsonx",
		BaseURL:    "https://us-south.ml.cloud.ibm.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "sap",
		Name:       "SAP AI Core",
		BaseURL:    "https://api.ai.core.cloud.sap",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "heroku",
		Name:       "Heroku AI",
		BaseURL:    "https://api.heroku.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})
}
