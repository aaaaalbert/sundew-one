package user

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	apps_v1alpha "edgenet/pkg/apis/apps/v1alpha"
	"edgenet/pkg/client/clientset/versioned"
	edgenettestclient "edgenet/pkg/client/clientset/versioned/fake"

	"github.com/Sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

//The main structure of test group
type UserTestGroup struct {
	authorityObj        apps_v1alpha.Authority
	authorityRequestObj apps_v1alpha.AuthorityRequest
	userObj             apps_v1alpha.User
	client              kubernetes.Interface
	edgenetclient       versioned.Interface
	handler             Handler
}

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	logrus.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

//Init syncs the test group
func (g *UserTestGroup) Init() {

	authorityObj := apps_v1alpha.Authority{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Authority",
			APIVersion: "apps.edgenet.io/v1alpha",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "edgenet",
		},
		Spec: apps_v1alpha.AuthoritySpec{
			FullName:  "EdgeNet",
			ShortName: "EdgeNet",
			URL:       "https://www.edge-net.org",
			Address: apps_v1alpha.Address{
				City:    "Paris - NY - CA",
				Country: "France - US",
				Street:  "4 place Jussieu, boite 169",
				ZIP:     "75005",
			},
			Contact: apps_v1alpha.Contact{
				Email:     "unittest@edge-net.org",
				FirstName: "unit",
				LastName:  "testing",
				Phone:     "+33NUMBER",
				Username:  "unittesting",
			},
		},
		Status: apps_v1alpha.AuthorityStatus{
			Enabled: false,
		},
	}
	authorityRequestObj := apps_v1alpha.AuthorityRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AuthorityRequest",
			APIVersion: "apps.edgenet.io/v1alpha",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "edgenet",
		},
		Spec: apps_v1alpha.AuthorityRequestSpec{
			FullName:  "EdgeNet",
			ShortName: "EdgeNet",
			URL:       "https://www.edge-net.org",
			Address: apps_v1alpha.Address{
				City:    "Paris",
				Country: "France",
				Street:  "4 place Jussieu, boite 169",
				ZIP:     "75005",
			},
			Contact: apps_v1alpha.Contact{
				Email:     "unittest@edge-net.org",
				FirstName: "unit",
				LastName:  "testing",
				Phone:     "+33NUMBER",
				Username:  "unittesting",
			},
		},
		Status: apps_v1alpha.AuthorityRequestStatus{
			State: success,
		},
	}

	userObj := apps_v1alpha.User{
		TypeMeta: metav1.TypeMeta{
			Kind:       "User",
			APIVersion: "apps.edgenet.io/v1alpha",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "unittesting",
		},
		Spec: apps_v1alpha.UserSpec{
			FirstName: "EdgeNet",
			LastName:  "EdgeNet",
			Roles:     []string{"Admin"},
			Email:     "unittest@edge-net.org",
		},
		Status: apps_v1alpha.UserStatus{
			State:  success,
			Active: true,
		},
	}

	g.authorityObj = authorityObj
	g.authorityRequestObj = authorityRequestObj
	g.userObj = userObj
	g.client = testclient.NewSimpleClientset()
	g.edgenetclient = edgenettestclient.NewSimpleClientset()
}

//TestHandlerInit for handler initialization
func TestHandlerInit(t *testing.T) {
	//Sync the test group
	g := UserTestGroup{}
	g.Init()
	// Initialize the handler
	g.handler.Init(g.client, g.edgenetclient)
	if g.handler.clientset != g.client {
		t.Error("Kubernetes clientset sync problem")
	}
	if g.handler.edgenetClientset != g.edgenetclient {
		t.Error("EdgeNet clientset sync problem")
	}
}

func TestUserCreate(t *testing.T) {
	g := UserTestGroup{}
	g.Init()
	g.handler.Init(g.client, g.edgenetclient)

	t.Run("creation of user ", func(t *testing.T) {
		g.handler.ObjectCreated(g.userObj.DeepCopy())
		user, _ := g.edgenetclient.AppsV1alpha().Users(fmt.Sprintf("users-%s", g.userObj.GetName())).Get(g.userObj.Spec.FirstName, metav1.GetOptions{})
		if user == nil {
			t.Error("User generation failed")
		}
	})
}

func TestGenerateRandomString(t *testing.T) {
	for i := 1; i < 5; i++ {
		origin := generateRandomString(16)
		test := generateRandomString(16)
		if origin == test {
			t.Error("User GenerateRadnomString failed")
		}
	}
}