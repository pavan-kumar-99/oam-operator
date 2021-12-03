package controllers

import (
	"context"
	appsv1beta1 "oam-operator/api/v1beta1"
	"time"

	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Application", func() {

	appnamespace := "default"
	appname := "bank-vault"
	objMeta := metav1.ObjectMeta{
		Namespace: appnamespace,
		Name:      appname,
		Labels:    map[string]string{"custom": "label"},
	}
	DeployName := types.NamespacedName{Name: appname, Namespace: appnamespace}
	app := &appsv1beta1.Application{
		ObjectMeta: objMeta,
		Spec: appsv1beta1.ApplicationSpec{
			ApplicationName: appname,
			Cloud: appsv1beta1.CloudSelector{
				Aws: appsv1beta1.AwsSpec{
					S3: "false",
				},
			},
		},
	}
	Context("when application is created with minimal configuration", func() {
		It("should be created a Deployment", func() {
			newApp := app.DeepCopy()
			err := k8sClient.Create(context.Background(), newApp)
			Expect(err).ToNot(HaveOccurred())
			deploy := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(context.Background(), DeployName, deploy)
			}, time.Second*5).Should(Succeed())
			Expect(deploy.Spec.Replicas).To(Equal(proto.Int32(2)))
		})
	})
})
