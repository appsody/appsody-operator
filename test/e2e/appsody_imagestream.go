package e2e

import (
	goctx "context"
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/test/util"
	imagev1 "github.com/openshift/api/image/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

//AppsodyImageStreamTest ...
func AppsodyImageStreamTest(t *testing.T) {
	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	f := framework.Global

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Couldn't get namespace: %v", err)
	}

	// Adds the imagestream resources to the scheme
	if err = imagev1.AddToScheme(f.Scheme); err != nil {
		t.Logf("Unable to add image scheme: (%v)", err)
		util.FailureCleanup(t, f, namespace, err)
	}

	t.Logf("Namespace: %s", namespace)

	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, timeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	if err = appsodyImageStreamTest(t, f, ctx); err != nil {
		out, deleteErr := exec.Command("oc", "delete", "imagestream", "imagestream-example").Output()
		if deleteErr != nil {
			t.Fatalf("Failed to delete imagestream: %s", out)
		}
		util.FailureCleanup(t, f, namespace, err)
	}
}

func appsodyImageStreamTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	const name = "appsody-app"
	const imgstreamName = "imagestream-example"

	ns, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	t.Logf("Namespace: %s", ns)
	target := types.NamespacedName{Name: name, Namespace: ns}

	// Create the imagestream
	out, err := exec.Command("oc", "import-image", imgstreamName, "--from=navidsh/demo-day:v0.1.0", "-n", ns, "--confirm").Output()
	if err != nil {
		t.Fatalf("Creating the imagestream failed: %s", out)
	}

	err = waitForImageStream(f, ctx, imgstreamName, ns)
	if err != nil {
		return err
	}

	// Make an appplication that points to the imagestream
	appsody := util.MakeBasicAppsodyApplication(t, f, name, ns, 1)
	appsody.Spec.ApplicationImage = imgstreamName

	timestamp := time.Now().UTC()
	t.Logf("%s - Creating appsody application...", timestamp)
	err = f.Client.Create(goctx.TODO(), appsody,
		&framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, ns, name, 1, retryInterval, timeout)
	if err != nil {
		return err
	}


	previousImage, err := getCurrImageRef(f, ctx, target)
	if err != nil {
		return err
	}
	// Update the imagestreamtag
	tag := `{"tag":{"from":{"name": "navidsh/demo-day:v0.2.0"}}}`
	out, err = exec.Command("oc", "patch", "imagestreamtag", imgstreamName+":latest", "-n", ns, "-p", tag).Output()
	if err != nil {
		t.Fatalf("Updating the imagestreamtag failed: %s", out)
	}

	// Return err if the image reference is not updated successfully
	err = waitImageRefUpdated(t, f, ctx, target, previousImage)
	if err != nil {
		return err
	}

	previousImage, err = getCurrImageRef(f, ctx, target)
	if err != nil {
		return err
	}
	// Update the imagestreamtag again
	tag = `{"tag":{"from":{"name": "navidsh/demo-day:v0.1.0"}}}`
	out, err = exec.Command("oc", "patch", "imagestreamtag", imgstreamName+":latest", "-n", ns, "-p", tag).Output()
	if err != nil {
		t.Fatalf("Updating the imagestreamtag failed: %s", out)
	}

	// Return err if the image reference is not updated successfully
	err = waitImageRefUpdated(t, f, ctx, target, previousImage)
	if err != nil {
		return err
	}

	out, err = exec.Command("oc", "delete", "imagestream", "imagestream-example", "-n", ns).Output()
	if err != nil {
		t.Fatalf("Failed to delete imagestream: %s", out)
	}

	if err = testRemoveImageStream(t, f, ctx); err != nil {
		t.Fatal(err)
	}

	return nil
}

func testRemoveImageStream(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	const name = "appsody-app"
	ns, err := ctx.GetNamespace()
	if err != nil {
		return err
	}
	target := types.NamespacedName{Namespace: ns, Name: name}
	err = util.UpdateApplication(f, target, func(r *appsodyv1beta1.AppsodyApplication) {
		r.Spec.ApplicationImage = "navidsh/demo-day"
	})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, ns, name, 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	imageRef, err := getCurrImageRef(f, ctx, target)
	if err != nil {
		return err
	}

	if imageRef != "navidsh/demo-day" {
		return errors.New("image reference not updated to docker hub ref")
	}

	return nil
}

/* Helper Functions Below */
func waitForImageStream(f *framework.Framework, ctx *framework.TestCtx, imgstreamName string, ns string) error {
	// Check the name field that matches
	key := map[string]string{"metadata.name": imgstreamName}

	options := &dynclient.ListOptions{
		FieldSelector: fields.Set(key).AsSelector(),
		Namespace:     ns,
	}

	imageStreamList := &imagev1.ImageStreamList{}

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = f.Client.List(goctx.TODO(), imageStreamList, options)
		if err != nil {
			return true, err
		}

		if len(imageStreamList.Items) == 0 {
			return false, nil
		}

		return true, nil
	})

	if errors.Is(err, wait.ErrWaitTimeout) {
		return errors.New("imagestream not found")
	}

	return err
}

func getCurrImageRef(f *framework.Framework, ctx *framework.TestCtx,
		target types.NamespacedName) (string, error) {
	appsody := appsodyv1beta1.AppsodyApplication{}
	err := f.Client.Get(goctx.TODO(), target, &appsody)
	if err != nil {
		return "", err
	}
	return appsody.Status.ImageReference, nil
}

func waitImageRefUpdated(t *testing.T, f *framework.Framework, ctx *framework.TestCtx,
		target types.NamespacedName, imageRef string) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		currImage, err := getCurrImageRef(f, ctx, target)
		if err != nil {
			return true, err	// if error, stop polling and return err
		}

		// Check if the image the application is pointing to has been changed
		if currImage == imageRef {
			// keep polling if the image ref is not updated
			t.Log("Waiting for the image reference to be updated ...")
			return false, nil
		}
		return true, nil
	})

	if errors.Is(err, wait.ErrWaitTimeout) {
		return errors.New("image reference not updated")
	}

	return err	// implicitly return nil if no errors
}