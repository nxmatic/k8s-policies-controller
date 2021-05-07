package k8s

import (
	"context"
	"strings"

	errors_api "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	gcpauth_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpauth/v1alpha1"
	gcpauth_policy_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpauth/v1alpha1"
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s_spi "github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/k8s"
)

type (
	Interface struct {
		k8s_spi.Interface
	}

	Profile = gcpauth_api.Profile
	Spec    = gcpauth_api.ProfileSpec
	Status  = gcpauth_api.ProfileStatus

	KeyValue = gcpauth_api.KeyValue
)

const (
	ProfileKey = gcpauth_api.ProfilesKey
)

var (
	ProfilesResource = gcpauth_api.ProfilesResource
	SecretsResource  = core_api.SchemeGroupVersion.WithResource("secrets")
)

func NewInterface(mgr manager.Manager) *Interface {
	spi := k8s_spi.NewInterface(mgr)
	return &Interface{
		Interface: *spi,
	}
}

func (s *Interface) ResolveProfile(meta meta_api.ObjectMeta) (*gcpauth_policy_api.Profile, error) {
	if resolved, err := s.Interface.ResolveProfile(meta, s.newProfileCollector()); err != nil {
		return nil, err
	} else {
		profile := resolved.(ProfileAdaptor)
		return profile.Profile, nil
	}
}
func (s *Interface) GetProfile(name string) (*Profile, error) {
	key := client.ObjectKey{
		Name: name,
	}
	profile := &Profile{}
	if err := s.Manager.GetClient().Get(context.Background(), key, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (i *Interface) GetDatasourceSecret(profile *Profile) (*core_api.Secret, error) {
	secret := core_api.Secret{
		ObjectMeta: meta_api.ObjectMeta{
			Name:      profile.Spec.Datasource.Name,
			Namespace: profile.Spec.Datasource.Namespace,
		},
	}
	return i.GetSecret(&secret)
}

func (i *Interface) GetSecret(secret *core_api.Secret) (*core_api.Secret, error) {
	key := client.ObjectKey{
		Name:      secret.ObjectMeta.Name,
		Namespace: secret.ObjectMeta.Namespace,
	}
	if err := i.Manager.GetClient().Get(context.Background(), key, secret); err != nil {
		return nil, err
	}

	return secret, nil
}

func (i *Interface) DeleteImagePullSecret(name string) error {
	return i.Manager.GetClient().DeleteAllOf(
		context.Background(),
		&core_api.Secret{},
		client.MatchingLabels{
			ProfileKey.String(): name,
		})
}

func (i *Interface) DeleteSecret(secret *core_api.Secret) error {
	return i.Manager.GetClient().Delete(context.Background(), secret)
}

func (i *Interface) GetImagePullSecret(profile *Profile) (*core_api.Secret, error) {
	return i.GetSecret(&core_api.Secret{
		ObjectMeta: meta_api.ObjectMeta{
			Name: profile.ObjectMeta.Name,
		},
	})
}

func (i *Interface) UpdateImagePullSecret(profile *Profile) error {
	var err error = nil

	datasource, err := i.GetDatasourceSecret(profile)
	if err != nil {
		return err
	}

	if secret, err := i.GetImagePullSecret(profile); err != nil {
		if errors_api.IsNotFound(err) {
			_, err = i.CreateImagePullSecret(profile, datasource)
		}
	} else {
		secret.Data[core_api.DockerConfigJsonKey] = datasource.Data[core_api.DockerConfigJsonKey]
		_, err = i.UpdateSecret(secret)
	}

	return err
}

func (i *Interface) CreateImagePullSecret(profile *Profile, datasource *core_api.Secret) (*core_api.Secret, error) {
	secret := &core_api.Secret{
		ObjectMeta: meta_api.ObjectMeta{
			Name: profile.ObjectMeta.Name,
			Labels: map[string]string{
				ProfileKey.String():           profile.ObjectMeta.Name,
				gcpauth_api.WatchKey.String(): "true",
			},
		},
		Type: core_api.SecretTypeDockerConfigJson,
		Data: map[string][]uint8{
			core_api.DockerConfigJsonKey: datasource.Data[core_api.DockerConfigJsonKey],
		},
	}
	return i.CreateSecret(secret)

}

func (i *Interface) CreateSecret(secret *core_api.Secret) (*core_api.Secret, error) {
	if err := i.Manager.GetClient().Create(context.Background(), secret); err != nil {
		return nil, err
	}

	return i.GetSecret(secret)
}

func (i *Interface) UpdateSecret(secret *core_api.Secret) (*core_api.Secret, error) {
	if err := i.Manager.GetClient().Update(context.Background(), secret); err != nil {
		return nil, err
	}
	return i.GetSecret(secret)
}

func (i *Interface) EnsureNamespaceImagePullSecret(profile *Profile, namespace string) error {
	secret, err := i.GetImagePullSecret(profile)
	if err != nil {
		return err
	}
	namespaces, ok := secret.ObjectMeta.Annotations["replicator.v1.mittwald.de/replicate-to"]
	if !ok {
		namespaces = namespace
	} else {
		if strings.Contains(namespaces, namespace) {
			return nil
		}
		namespaces = namespaces + "," + namespace
	}
	if secret.ObjectMeta.Annotations == nil {
		secret.ObjectMeta.Annotations = make(map[string]string)
	}
	secret.ObjectMeta.Annotations["replicator.v1.mittwald.de/replicate-to"] = namespaces
	_, err = i.UpdateSecret(secret)
	return err
}

func (k8s Interface) newProfileCollector() k8s_spi.ProfileCollector {
	return k8s_spi.NewProfileCollector(
		k8s.newProfileSupplier())
}

func (k8s Interface) newProfileSupplier() k8s_spi.ProfileSupplier {
	return ProfileSupplier{
		k8s,
	}
}
