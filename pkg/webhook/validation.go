package webhook

import (
	"context"
	admission "k8s.io/api/admissionregistration/v1"
	api "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/admissionregistration/v1"
	meta "k8s.io/client-go/applyconfigurations/meta/v1"

	"k8s.io/client-go/kubernetes"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"vcluster-gatekeeper/cert"
)

func ApplyValidationConfig(crt *cert.Certificate) {

	var (
		webhookNamespace, _ = os.LookupEnv("WEBHOOK_NAMESPACE")
		// validationCfgName, _ = os.LookupEnv("VALIDATE_CONFIG") Not used here in below code
		webhookService, _ = os.LookupEnv("WEBHOOK_SERVICE")
	)
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic("failed to set go -client")
	}

	path := "/validate"
	fail := admission.Fail
	sideEffect := admission.SideEffectClassNone
	scope := admission.ClusterScope
	webhookName := webhookService + "." + webhookNamespace + ".svc.cluster.local"
	kind := "ValidatingWebhookConfiguration"
	apiVersion := "admissionregistration.k8s.io/v1"

	validateConfig := &v1.ValidatingWebhookConfigurationApplyConfiguration{
		TypeMetaApplyConfiguration: meta.TypeMetaApplyConfiguration{
			Kind:       &kind,
			APIVersion: &apiVersion,
		},
		ObjectMetaApplyConfiguration: &meta.ObjectMetaApplyConfiguration{
			Name: &webhookService,
		},
		Webhooks: []v1.ValidatingWebhookApplyConfiguration{{
			Name:                    &webhookName,
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &sideEffect,
			ClientConfig: &v1.WebhookClientConfigApplyConfiguration{
				CABundle: crt.CaCert.Bytes(), // CA bundle created earlier
				Service: &v1.ServiceReferenceApplyConfiguration{
					Name:      &webhookService,
					Namespace: &webhookNamespace,
					Path:      &path,
				},
			},
			Rules: []v1.RuleWithOperationsApplyConfiguration{{Operations: []admission.OperationType{
				admission.Delete},
				RuleApplyConfiguration: v1.RuleApplyConfiguration{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"namespaces"},
					Scope:       &scope,
				},
			}},
			FailurePolicy: &fail,
		}},
	}

	if _, err := kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Apply(context.Background(),
		validateConfig,
		api.ApplyOptions{FieldManager: "vcluster-gatekeeper", Force: true}); err != nil {
		panic(err)
	}
}
