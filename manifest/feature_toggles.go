package manifest

type FeatureToggles []string

const FeatureUpdatePipeline = "update-pipeline"
const FeatureDockerDecompose = "docker-decompose"
const FeatureCFV7 = "cf-v7"
const FeatureNewDeployResource = "new-deploy-resource"
const FeatureToggleDisableDeprecatedDockerRegistryError = "im-aware-that-old-docker-registries-will-stop-working-on-24-august-2020"

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureDockerDecompose,
	FeatureCFV7,
	FeatureNewDeployResource,
	FeatureToggleDisableDeprecatedDockerRegistryError,
}

func (f FeatureToggles) contains(aFeature string) bool {
	for _, feature := range f {
		if feature == aFeature {
			return true
		}
	}
	return false
}

func (f FeatureToggles) Versioned() bool {
	return f.UpdatePipeline()
}

func (f FeatureToggles) NewDeployResource() bool {
	return f.contains(FeatureNewDeployResource)
}

func (f FeatureToggles) UpdatePipeline() bool {
	return f.contains(FeatureUpdatePipeline)
}

func (f FeatureToggles) DockerDecompose() bool {
	return f.contains(FeatureDockerDecompose)
}

func (f FeatureToggles) CFV7() bool {
	return f.contains(FeatureCFV7)
}

func (f FeatureToggles) DisableDockerRegistryLinter() bool {
	return f.contains(FeatureToggleDisableDeprecatedDockerRegistryError)
}
