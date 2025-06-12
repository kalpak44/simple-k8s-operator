package controller

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	homev1 "github.com/kalpak44/simple-k8s-operator/api/v1"
)

var _ = Describe("Backup Controller", func() {
	var (
		ctx            context.Context
		namespacedName types.NamespacedName
		backup         *homev1.Backup
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespacedName = types.NamespacedName{
			Namespace: "default",
			Name:      "feature-test",
		}

		// Создаём Backup-ресурс
		backup = &homev1.Backup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			},
			Spec: homev1.BackupSpec{
				Database: "test-db",
				Schedule: "*/5 * * * *",
			},
		}
		Expect(k8sClient.Create(ctx, backup)).To(Succeed())
	})

	AfterEach(func() {
		// Чистим Backup и соответствующий CronJob
		Expect(k8sClient.Delete(ctx, backup)).To(Succeed())

		// Ждём, пока CronJob исчезнет
		cron := &batchv1.CronJob{}
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{
				Namespace: namespacedName.Namespace,
				Name:      namespacedName.Name + "-cron",
			}, cron)
		}, "5s", "500ms").Should(HaveOccurred())
	})

	It("should create a CronJob with correct spec and ownerReference", func() {
		// Вызываем Reconcile
		reconciler := &BackupReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
		Expect(err).NotTo(HaveOccurred())

		// Проверяем, что CronJob появился
		cron := &batchv1.CronJob{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Namespace: namespacedName.Namespace,
			Name:      namespacedName.Name + "-cron",
		}, cron)).To(Succeed())

		// 1) Schedule совпадает
		Expect(cron.Spec.Schedule).To(Equal(backup.Spec.Schedule))

		// 2) Container с curl и правильными аргументами
		containers := cron.Spec.JobTemplate.Spec.Template.Spec.Containers
		Expect(containers).To(HaveLen(1))
		c := containers[0]
		Expect(c.Image).To(Equal("curlimages/curl:latest"))
		Expect(c.Args).To(ContainElements(
			"-s",
			"-X", "POST",
			"https://kalpak44.free.beeceptor.com",
			fmt.Sprintf("db=%s", backup.Spec.Database),
		))

		// 3) OwnerReference указывает на Backup
		Expect(len(cron.OwnerReferences)).To(Equal(1))
		or := cron.OwnerReferences[0]
		Expect(or.Kind).To(Equal("Backup"))
		Expect(or.Name).To(Equal(backup.Name))
	})
})
