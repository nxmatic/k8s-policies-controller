package k8s

import (
	"context"
	"strings"

	errors_api "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/client-go/dynamic"

	gcpauth_api "github.com/nuxeo/k8s-policy-controller/apis/gcpauthpolicyprofile/v1alpha1"
	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s_spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"
	"k8s.io/apimachinery/pkg/runtime"
)

type (
	Interface struct {
		*k8s_spi.Interface
	}

	Profile = gcpauth_api.Profile
	Spec    = gcpauth_api.ProfileSpec
	Status  = gcpauth_api.ProfileStatus

	TypeValue = gcpauth_api.TypeValue
	KeyValue  = gcpauth_api.KeyValue
)

const (
	ProfileKey               = gcpauth_api.ProfileKey
	ImagePullSecretTypeValue = gcpauth_api.ImagePullSecretTypeValue
)

var (
	ProfilesResource = gcpauth_api.ProfilesResource
	SecretsResource  = core_api.SchemeGroupVersion.WithResource("secrets")
)

func NewInterface(client dynamic.Interface) *Interface {
	return &Interface{
		k8s_spi.NewInterface(client),
	}
}

func (s *Interface) ResolveProfile(namespace *meta_api.ObjectMeta, resource *meta_api.ObjectMeta) (*gcpauth_api.Profile, error) {
	annotations := make(map[string]string)
	annotations = s.MergeAnnotations(annotations, namespace)
	annotations = s.MergeAnnotations(annotations, resource)

	if names, ok := annotations[gcpauth_api.ProfileKey.String()]; ok {
		for _, name := range strings.Split(names, ",") {
			profile, err := s.GetProfile(name)
			if err != nil {
				return nil, errors.New("cannot retrieve profile " + name)
			}
			if profile.Spec.Selector != nil {
				selector, err := meta_api.LabelSelectorAsSelector(profile.Spec.Selector)
				if err != nil {
					return nil, err
				}
				if !selector.Matches(labels.Set(resource.Labels)) {
					continue
				}
			}
			return profile, nil
		}
	}
	return nil, errors.New("Annotation not found")

}

func (s *Interface) GetProfile(name string) (*Profile, error) {
	resp, err := s.Interface.Resource(ProfilesResource).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	profile := &Profile{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), profile)
	if err != nil {
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
	resp, err := i.Interface.
		Resource(SecretsResource).
		Namespace(secret.ObjectMeta.Namespace).
		Get(context.TODO(), secret.ObjectMeta.Name,
			meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (i *Interface) DeleteImagePullSecret(name string) error {
	return i.Resource(SecretsResource).
		DeleteCollection(context.TODO(),
			meta_api.DeleteOptions{},
			meta_api.ListOptions{
				LabelSelector: labels.Set{
					ProfileKey.String(): name,
				}.String(),
			})
}

func (i *Interface) DeleteSecret(secret *core_api.Secret) error {
	return i.Resource(SecretsResource).
		Namespace(secret.ObjectMeta.Namespace).
		Delete(context.TODO(), secret.ObjectMeta.Name, meta_api.DeleteOptions{})
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
	secret := core_api.Secret{
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

	return i.CreateSecret(&secret)
}

func (i *Interface) CreateSecret(secret *core_api.Secret) (*core_api.Secret, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
	if err != nil {
		return nil, err
	}
	resp, err :=
		i.Resource(SecretsResource).
			Namespace(secret.ObjectMeta.Namespace).
			Create(context.TODO(), &unstructured.Unstructured{Object: data}, meta_api.CreateOptions{})
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (i *Interface) UpdateSecret(secret *core_api.Secret) (*core_api.Secret, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
	if err != nil {
		return nil, err
	}
	resp, err :=
		i.Resource(SecretsResource).
			Namespace(secret.ObjectMeta.Namespace).
			Update(context.TODO(), &unstructured.Unstructured{Object: data}, meta_api.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
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
	secret, err = i.UpdateSecret(secret)
	return err
}
