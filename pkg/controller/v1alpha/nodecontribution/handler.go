/*
Copyright 2020 Sorbonne Université

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodecontribution

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	apps_v1alpha "edgenet/pkg/apis/apps/v1alpha"
	"edgenet/pkg/client/clientset/versioned"
	"edgenet/pkg/controller/v1alpha/authority"
	"edgenet/pkg/mailer"
	ns "edgenet/pkg/namespace"
	"edgenet/pkg/node"
	"edgenet/pkg/remoteip"

	log "github.com/Sirupsen/logrus"
	namecheap "github.com/billputer/go-namecheap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HandlerInterface interface contains the methods that are required
type HandlerInterface interface {
	Init(kubernetes kubernetes.Interface, edgenet versioned.Interface) error
	ObjectCreated(obj interface{})
	ObjectUpdated(obj interface{})
	ObjectDeleted(obj interface{})
}

// Handler implementation
type Handler struct {
	clientset        kubernetes.Interface
	edgenetClientset versioned.Interface
	publicKey        ssh.Signer
}

// Init handles any handler initialization
func (t *Handler) Init(kubernetes kubernetes.Interface, edgenet versioned.Interface) error {
	log.Info("NCHandler.Init")
	t.clientset = kubernetes
	t.edgenetClientset = edgenet

	// Get the SSH Public Key of the headnode
	key, err := ioutil.ReadFile("../../.ssh/id_rsa")
	if err != nil {
		log.Println(err.Error())
		panic(err.Error())
	}

	t.publicKey, err = ssh.ParsePrivateKey(key)
	if err != nil {
		log.Println(err.Error())
		panic(err.Error())
	}
	node.Clientset = t.clientset
	return err
}

// ObjectCreated is called when an object is created
func (t *Handler) ObjectCreated(obj interface{}) {
	log.Info("NCHandler.ObjectCreated")
	// Create a copy of the node contribution object to make changes on it
	NCCopy := obj.(*apps_v1alpha.NodeContribution).DeepCopy()
	NCCopy.Status.Message = []string{}
	// Find the authority from the namespace in which the object is
	NCOwnerNamespace, _ := t.clientset.CoreV1().Namespaces().Get(NCCopy.GetNamespace(), metav1.GetOptions{})
	nodeName := fmt.Sprintf("%s.%s.edge-net.io", NCOwnerNamespace.Labels["authority-name"], NCCopy.GetName())
	// Don't use the authority name if the node belongs to EdgeNet
	if NCOwnerNamespace.GetName() == "authority-edgenet" {
		nodeName = fmt.Sprintf("%s.edge-net.io", NCCopy.GetName())
	}
	NCOwnerAuthority, _ := t.edgenetClientset.AppsV1alpha().Authorities().Get(NCOwnerNamespace.Labels["authority-name"], metav1.GetOptions{})
	authorityEnabled := NCOwnerAuthority.Spec.Enabled
	log.Println("AUTHORITY CHECK")
	// Check if the authority is active
	if authorityEnabled {
		log.Println("AUTHORITY ENABLED")
		// If the service restarts, it creates all objects again
		// Because of that, this section covers a variety of possibilities
		// Check whether the host has been given as an IP address or else
		recordType := remoteip.GetRecordType(NCCopy.Spec.Host)
		if recordType == "" {
			NCCopy.Status.State = failure
			NCCopy.Status.Message = append(NCCopy.Status.Message, statusDict["invalid-host"])
			t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
			t.sendEmail(NCCopy)
			return
		}
		// Set the client config according to the node contribution,
		// with the maximum time of 15 seconds to establist the connection.
		config := &ssh.ClientConfig{
			User:            NCCopy.Spec.User,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(t.publicKey), ssh.Password(NCCopy.Spec.Password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         15 * time.Second,
		}
		addr := fmt.Sprintf("%s:%d", NCCopy.Spec.Host, NCCopy.Spec.Port)
		contributedNode, err := t.clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
		if err == nil {
			// The node corresponding to the contributed node exists in the cluster
			log.Println("NODE FOUND")
			if node.GetConditionReadyStatus(contributedNode.DeepCopy()) != trueStr {
				go t.runRecoveryProcedure(addr, config, nodeName, NCCopy, contributedNode)
			} else {
				NCCopy.Status.State = success
				NCCopy.Status.Message = append(NCCopy.Status.Message, statusDict["node-ok"])
				t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
			}
		} else {
			// There isn't any node corresponding to the node contribution
			log.Println("NODE NOT FOUND")
			go t.runSetupProcedure(NCOwnerNamespace.Labels["authority-name"], addr, nodeName, recordType, config, NCCopy)
		}
	} else {
		log.Println("AUTHORITY NOT ENABLED")
		// Disable scheduling on the node if the authority is disabled
		NCCopy.Spec.Enabled = false
		NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).Update(NCCopy)
		if err == nil {
			NCCopy = NCCopyUpdated
			NCCopy.Status.State = failure
			NCCopy.Status.Message = append(NCCopy.Status.Message, statusDict["authority-disabled"])
			t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
		}
	}
}

// ObjectUpdated is called when an object is updated
func (t *Handler) ObjectUpdated(obj interface{}) {
	log.Info("NCHandler.ObjectUpdated")
	// Create a copy of the node contribution object to make changes on it
	NCCopy := obj.(*apps_v1alpha.NodeContribution).DeepCopy()
	NCCopy.Status.Message = []string{}

	NCOwnerNamespace, _ := t.clientset.CoreV1().Namespaces().Get(NCCopy.GetNamespace(), metav1.GetOptions{})
	nodeName := fmt.Sprintf("%s.%s.edge-net.io", NCOwnerNamespace.Labels["authority-name"], NCCopy.GetName())
	var authorityEnabled bool
	if NCOwnerNamespace.GetName() == "authority-edgenet" {
		nodeName = fmt.Sprintf("%s.edge-net.io", NCCopy.GetName())
		authorityEnabled = true
	} else {
		NCOwnerAuthority, _ := t.edgenetClientset.AppsV1alpha().Authorities().Get(NCOwnerNamespace.Labels["authority-name"], metav1.GetOptions{})
		authorityEnabled = NCOwnerAuthority.Spec.Enabled
	}
	log.Println("AUTHORITY CHECK")
	// Check if the authority is active
	if authorityEnabled {
		log.Println("AUTHORITY ENABLED")
		recordType := remoteip.GetRecordType(NCCopy.Spec.Host)
		if recordType == "" {
			NCCopy.Status.State = failure
			NCCopy.Status.Message = append(NCCopy.Status.Message, statusDict["invalid-host"])
			t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
			t.sendEmail(NCCopy)
			return
		}
		config := &ssh.ClientConfig{
			User:            NCCopy.Spec.User,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(t.publicKey), ssh.Password(NCCopy.Spec.Password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         15 * time.Second,
		}
		addr := fmt.Sprintf("%s:%d", NCCopy.Spec.Host, NCCopy.Spec.Port)
		contributedNode, err := t.clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
		if err == nil {
			log.Println("NODE FOUND")
			if contributedNode.Spec.Unschedulable != !NCCopy.Spec.Enabled {
				node.SetNodeScheduling(nodeName, !NCCopy.Spec.Enabled)
			}
			if NCCopy.Status.State == failure {
				go t.runRecoveryProcedure(addr, config, nodeName, NCCopy, contributedNode)
			}
		} else {
			log.Println("NODE NOT FOUND")
			go t.runSetupProcedure(NCOwnerNamespace.Labels["authority-name"], addr, nodeName, recordType, config, NCCopy)
		}
	} else {
		log.Println("AUTHORITY NOT ENABLED")
		NCCopy.Spec.Enabled = false
		NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).Update(NCCopy)
		if err == nil {
			NCCopy = NCCopyUpdated
			NCCopy.Status.State = failure
			NCCopy.Status.Message = append(NCCopy.Status.Message, "Authority disabled")
			t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
		}
	}
}

// ObjectDeleted is called when an object is deleted
func (t *Handler) ObjectDeleted(obj interface{}) {
	log.Info("NCHandler.ObjectDeleted")
	// Mail notification, TBD
}

// sendEmail to send notification to participants
func (t *Handler) sendEmail(NCCopy *apps_v1alpha.NodeContribution) {
	// For those who are authority-admin and authorized users of the authority
	userRaw, err := t.edgenetClientset.AppsV1alpha().Users(NCCopy.GetNamespace()).List(metav1.ListOptions{})
	if err == nil {
		contentData := mailer.MultiProviderData{}
		contentData.Name = NCCopy.GetName()
		contentData.Host = NCCopy.Spec.Host
		contentData.Status = NCCopy.Status.State
		contentData.Message = NCCopy.Status.Message
		for _, userRow := range userRaw.Items {
			if userRow.Spec.Active && userRow.Status.AUP && userRow.Status.Type == "admin" {
				if err == nil && userRow.Spec.Active && userRow.Status.AUP {
					// Set the HTML template variables
					contentData.CommonData.Authority = userRow.GetNamespace()
					contentData.CommonData.Username = userRow.GetName()
					contentData.CommonData.Name = fmt.Sprintf("%s %s", userRow.Spec.FirstName, userRow.Spec.LastName)
					contentData.CommonData.Email = []string{userRow.Spec.Email}
					if contentData.Status == failure {
						mailer.Send("node-contribution-failure", contentData)
					} else if contentData.Status == success {
						mailer.Send("node-contribution-successful", contentData)
					}
				}
			}
		}
		if contentData.Status == failure {
			mailer.Send("node-contribution-failure-support", contentData)
		}
	}
}

// runSetupProcedure installs necessary packages from scratch and makes the node join into the cluster
func (t *Handler) runSetupProcedure(authorityName, addr, nodeName, recordType string, config *ssh.ClientConfig,
	NCCopy *apps_v1alpha.NodeContribution) error {
	// Steps in the procedure
	endProcedure := make(chan bool, 1)
	dnsConfiguration := make(chan bool, 1)
	installation := make(chan bool, 1)
	nodePatch := make(chan bool, 1)
	// Set the status as recovering
	NCCopy.Status.State = inprogress
	NCCopy.Status.Message = append(NCCopy.Status.Message, "Installation procedure has started")
	NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
	if err == nil {
		NCCopy = NCCopyUpdated
	}
	// Start DNS configuration of `edge-net.io`
	dnsConfiguration <- true
	// This statement to organize tasks and put a general timeout on
nodeInstallLoop:
	for {
		select {
		case <-dnsConfiguration:
			log.Println("***************DNS Configuration***************")
			// Use Namecheap API for registration
			hostRecord := namecheap.DomainDNSHost{
				Name:    strings.TrimSuffix(nodeName, ".edge-net.io"),
				Type:    recordType,
				Address: NCCopy.Spec.Host,
			}
			result, state := node.SetHostname(hostRecord)
			// If the host record already exists, update the status of the node contribution.
			// However, the setup procedure keeps going on, so, it is not terminated.
			if !result {
				var hostnameError string
				if state == "exist" {
					hostnameError = fmt.Sprintf("Error: Hostname %s or address %s already exists", hostRecord.Name, hostRecord.Address)
				} else {
					hostnameError = fmt.Sprintf("Error: Hostname %s or address %s couldn't added", hostRecord.Name, hostRecord.Address)
				}
				NCCopy.Status.State = incomplete
				NCCopy.Status.Message = append(NCCopy.Status.Message, hostnameError)
				NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
				if err == nil {
					NCCopy = NCCopyUpdated
				}
				log.Println(hostnameError)
			}
			installation <- true
		case <-installation:
			log.Println("***************Installation***************")
			// To prevent hanging forever during establishing a connection
			go func() {
				// SSH into the node
				conn, err := ssh.Dial("tcp", addr, config)
				if err != nil {
					log.Println(err)
					NCCopy.Status.State = failure
					NCCopy.Status.Message = append(NCCopy.Status.Message, "SSH handshake failed")
					NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
					log.Println(err)
					if err == nil {
						NCCopy = NCCopyUpdated
					}
					endProcedure <- true
					return
				}
				defer conn.Close()
				// Uninstall all existing packages related, do a clean installation, and make the node join to the cluster
				err = t.cleanInstallation(conn, nodeName, NCCopy)
				if err != nil {
					NCCopy.Status.State = failure
					NCCopy.Status.Message = append(NCCopy.Status.Message, "Node installation failed")
					NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
					log.Println(err)
					if err == nil {
						NCCopy = NCCopyUpdated
					}
					endProcedure <- true
					return
				}
				_, err = t.clientset.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
				if err == nil {
					nodePatch <- true
				}
			}()
		case <-nodePatch:
			log.Println("***************Node Patch***************")
			// Set the node as schedulable or unschedulable according to the node contribution
			patchStatus := true
			err := node.SetNodeScheduling(nodeName, !NCCopy.Spec.Enabled)
			if err != nil {
				NCCopy.Status.State = incomplete
				NCCopy.Status.Message = append(NCCopy.Status.Message, "Scheduling configuration failed")
				t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
				t.sendEmail(NCCopy)
				patchStatus = false
			}
			var ownerReferences []metav1.OwnerReference
			authorityCopy, err := t.edgenetClientset.AppsV1alpha().Authorities().Get(authorityName, metav1.GetOptions{})
			if err == nil {
				ownerReferences = authority.SetAsOwnerReference(authorityCopy)
			}
			NCOwnerNamespace, err := t.clientset.CoreV1().Namespaces().Get(fmt.Sprintf("authority-%s", authorityName), metav1.GetOptions{})
			if err == nil {
				ownerReferences = append(ownerReferences, ns.SetAsOwnerReference(NCOwnerNamespace)...)
			}
			err = node.SetOwnerReferences(nodeName, ownerReferences)
			if err != nil {
				NCCopy.Status.State = incomplete
				NCCopy.Status.Message = append(NCCopy.Status.Message, "Setting owner reference failed")
				t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
				t.sendEmail(NCCopy)
				patchStatus = false
			}
			if patchStatus {
				break nodeInstallLoop
			}
			NCCopy.Status.State = success
			NCCopy.Status.Message = append(NCCopy.Status.Message, "Node installation successful")
			t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
			endProcedure <- true
		case <-endProcedure:
			log.Println("***************Procedure Terminated***************")
			t.sendEmail(NCCopy)
			break nodeInstallLoop
		case <-time.After(25 * time.Minute):
			log.Println("***************Timeout***************")
			// Terminate the procedure after 25 minutes
			NCCopy.Status.State = failure
			NCCopy.Status.Message = append(NCCopy.Status.Message, "Node installation failed: timeout")
			NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
			log.Println(err)
			if err == nil {
				NCCopy = NCCopyUpdated
			}
			t.sendEmail(NCCopy)
			break nodeInstallLoop
		}
	}
	return err
}

// runRecoveryProcedure applies predefined methods to recover the node
func (t *Handler) runRecoveryProcedure(addr string, config *ssh.ClientConfig,
	nodeName string, NCCopy *apps_v1alpha.NodeContribution, contributedNode *corev1.Node) {
	// Steps in the procedure
	endProcedure := make(chan bool, 1)
	establishConnection := make(chan bool, 1)
	reconfiguration := make(chan bool, 1)
	installation := make(chan bool, 1)
	reboot := make(chan bool, 1)
	// Set the status as recovering
	NCCopy.Status.State = recover
	NCCopy.Status.Message = append(NCCopy.Status.Message, "Node recovering")
	NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
	if err == nil {
		NCCopy = NCCopyUpdated
	}
	// Watch the events of node object
	watchNode, err := t.clientset.CoreV1().Nodes().Watch(metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name==%s", contributedNode.GetName())})
	if err == nil {
		go func() {
			// Get events from watch interface
			for nodeEvent := range watchNode.ResultChan() {
				// Get updated node object
				updatedNode, status := nodeEvent.Object.(*corev1.Node)
				if status {
					if nodeEvent.Type == "DELETED" {
						endProcedure <- true
					}

					if node.GetConditionReadyStatus(updatedNode) == trueStr {
						NCCopy.Status.State = success
						NCCopy.Status.Message = append([]string{}, "Node recovery successful")
						NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
						log.Println(err)
						if err == nil {
							NCCopy = NCCopyUpdated
						}
						endProcedure <- true
					}
				}
			}
		}()
	} else {
		// In case of any malfunction of watching node resources,
		// terminate the function
		endProcedure <- true
	}

	var conn *ssh.Client
	go func() {
		conn, err = ssh.Dial("tcp", addr, config)
		if err != nil {
			log.Println(err)
			NCCopy.Status.State = failure
			NCCopy.Status.Message = append(NCCopy.Status.Message, "Node recovery failed: SSH handshake failed")
			NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
			log.Println(err)
			if err == nil {
				NCCopy = NCCopyUpdated
			}
			endProcedure <- true
		} else {
			reconfiguration <- true
		}
	}()

	// connCounter to try establishing a connection for several times when the node is rebooted
	connCounter := 0

	// This statement to organize tasks and put a general timeout on
nodeRecoveryLoop:
	for {
		select {
		case <-establishConnection:
			log.Printf("***************Establish Connection***************%s", nodeName)
			go func() {
				// SSH into the node
				conn, err = ssh.Dial("tcp", addr, config)
				if err != nil && connCounter < 3 {
					log.Println(err)
					// Wait three minutes to try establishing a connection again
					time.Sleep(3 * time.Minute)
					establishConnection <- true
					connCounter++
				} else if err != nil && connCounter >= 3 {
					NCCopy.Status.State = failure
					NCCopy.Status.Message = append(NCCopy.Status.Message, "Node recovery failed: SSH handshake failed")
					NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
					log.Println(err)
					if err == nil {
						NCCopy = NCCopyUpdated
					}
					<-endProcedure
					return
				}
				installation <- true
			}()
		case <-reconfiguration:
			log.Printf("***************Reconfiguration***************%s", nodeName)
			// Restart Docker & Kubelet and flush iptables
			err = reconfigureNode(conn, contributedNode.GetName())
			if err != nil {
				NCCopy.Status.Message = append(NCCopy.Status.Message, "Node recovery failed: reconfiguration step")
				NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
				log.Println(err)
				if err == nil {
					NCCopy = NCCopyUpdated
				}
			}
			time.Sleep(3 * time.Minute)
			reboot <- true
		case <-installation:
			log.Println("***************Installation***************")
			// Uninstall all existing packages related, do a clean installation, and make the node join to the cluster
			err := t.cleanInstallation(conn, nodeName, NCCopy)
			if err != nil {
				NCCopy.Status.State = failure
				NCCopy.Status.Message = append(NCCopy.Status.Message, "Node recovery failed: installation step")
				NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
				log.Println(err)
				if err == nil {
					NCCopy = NCCopyUpdated
				}
				t.sendEmail(NCCopy)
				watchNode.Stop()
				break nodeRecoveryLoop
			}
		case <-reboot:
			log.Println("***************Reboot***************")
			// Reboot the node in a minute
			err = rebootNode(conn)
			if err != nil {
				NCCopy.Status.Message = append(NCCopy.Status.Message, "Node recovery failed: reboot step")
				NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
				log.Println(err)
				if err == nil {
					NCCopy = NCCopyUpdated
				}
			}
			conn.Close()
			time.Sleep(3 * time.Minute)
			establishConnection <- true
		case <-endProcedure:
			log.Println("***************Procedure Terminated***************")
			t.sendEmail(NCCopy)
			watchNode.Stop()
			break nodeRecoveryLoop
		case <-time.After(25 * time.Minute):
			log.Println("***************Timeout***************")
			// Terminate the procedure after 25 minutes
			NCCopy.Status.State = failure
			NCCopy.Status.Message = append(NCCopy.Status.Message, "Node recovery failed: timeout")
			NCCopyUpdated, err := t.edgenetClientset.AppsV1alpha().NodeContributions(NCCopy.GetNamespace()).UpdateStatus(NCCopy)
			log.Println(err)
			if err == nil {
				NCCopy = NCCopyUpdated
			}
			t.sendEmail(NCCopy)
			watchNode.Stop()
			break nodeRecoveryLoop
		}
	}
	if conn != nil {
		conn.Close()
	}
}

// cleanInstallation gets and runs the uninstallation and installation commands prepared
func (t *Handler) cleanInstallation(conn *ssh.Client, nodeName string, NCCopy *apps_v1alpha.NodeContribution) error {
	uninstallationCommands, err := getUninstallCommands(conn)
	if err != nil {
		log.Println(err)
		return err
	}
	installationCommands, err := getInstallCommands(conn, nodeName, node.GetKubeletVersion()[1:])
	if err != nil {
		log.Println(err)
		return err
	}
	// Have root privileges
	commands := append([]string{"sudo su"}, uninstallationCommands...)
	commands = append(commands, installationCommands...)
	sess, err := startSession(conn)
	if err != nil {
		log.Println(err)
		return err
	}
	defer sess.Close()
	// StdinPipe for commands
	stdin, err := sess.StdinPipe()
	if err != nil {
		log.Println(err)
		return err
	}
	//sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	sess, err = startShell(sess)
	if err != nil {
		log.Println(err)
		return err
	}
	// Run commands sequentially
	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	stdin.Close()
	// Wait for session to finish
	err = sess.Wait()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// rebootNode restarts node after a minute
func rebootNode(conn *ssh.Client) error {
	sess, err := startSession(conn)
	if err != nil {
		log.Println(err)
		return err
	}
	defer sess.Close()
	err = sess.Run("sudo shutdown -r +1")
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// reconfigureNode gets and runs the configuration commands prepared
func reconfigureNode(conn *ssh.Client, hostname string) error {
	configurationCommands, err := getReconfigurationCommands(conn, hostname)
	if err != nil {
		log.Println(err)
		return err
	}
	// Have root privileges
	commands := append([]string{"sudo su"}, configurationCommands...)
	sess, err := startSession(conn)
	if err != nil {
		log.Println(err)
		return err
	}
	defer sess.Close()
	// StdinPipe for commands
	stdin, err := sess.StdinPipe()
	if err != nil {
		log.Println(err)
		return err
	}
	//sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	sess, err = startShell(sess)
	if err != nil {
		log.Println(err)
		return err
	}
	// Run commands sequentially
	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	stdin.Close()
	// Wait for session to finish
	err = sess.Wait()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Start a new session in the connection
func startSession(conn *ssh.Client) (*ssh.Session, error) {
	sess, err := conn.NewSession()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return sess, nil
}

// Start a shell in the session
func startShell(sess *ssh.Session) (*ssh.Session, error) {
	// Start remote shell
	if err := sess.Shell(); err != nil {
		log.Println(err)
		return nil, err
	}
	return sess, nil
}

// getInstallCommands prepares the commands necessary according to the OS
func getInstallCommands(conn *ssh.Client, hostname string, kubernetesVersion string) ([]string, error) {
	sess, err := startSession(conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer sess.Close()
	// Detect the node OS
	output, err := sess.Output("cat /etc/os-release")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	commandsDependent := []string{}
	if ubuntuOrDebian, _ := regexp.MatchString("ID=\"ubuntu\".*|ID=ubuntu.*|ID=\"debian\".*|ID=debian.*", string(output[:])); ubuntuOrDebian {
		// The commands including kubernetes & docker installation for Ubuntu, and also kubeadm join command

		var ubuntuCommands = []string{
			fmt.Sprintf("apt-get install docker.io kubeadm=%[1]s-00 kubectl=%[1]s-00 kubelet=%[1]s-00 kubernetes-cni -y", kubernetesVersion),
		}
		var debianCommands = []string{
			"apt-get install software-properties-common -y",
			"curl -fsSL https://download.docker.com/linux/debian/gpg | sudo apt-key add -",
			"add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable\"",
			"apt-get update",
			fmt.Sprintf("apt-get install docker-ce kubeadm=%[1]s-00 kubectl=%[1]s-00 kubelet=%[1]s-00 kubernetes-cni -y", kubernetesVersion),
		}
		if ubuntu, _ := regexp.MatchString("ID=\"ubuntu\".*|ID=ubuntu.*", string(output[:])); ubuntu {
			commandsDependent = ubuntuCommands
		} else {
			commandsDependent = debianCommands
		}
		commands1 := []string{
			"dpkg --configure -a",
			"apt-get update -y && apt-get install -y apt-transport-https -y",
			"apt-get install curl -y",
			"modprobe br_netfilter",
			"cat <<EOF > /etc/sysctl.d/k8s.conf",
			"net.bridge.bridge-nf-call-ip6tables = 1",
			"net.bridge.bridge-nf-call-iptables = 1",
			"EOF",
			"sysctl --system",
			"swapoff -a",
			"sed -e '/swap/ s/^#*/#/' -i /etc/fstab",
			"curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -",
			"cat <<EOF | tee /etc/apt/sources.list.d/kubernetes.list",
			"deb https://apt.kubernetes.io/ kubernetes-xenial main",
			"EOF",
			"apt-get update",
		}
		commands2 := []string{
			"apt-mark hold kubelet kubeadm kubectl",
			fmt.Sprintf("hostname %s", hostname),
			"systemctl enable docker",
			"systemctl start docker",
			node.CreateJoinToken("600s", hostname),
			"systemctl daemon-reload",
			"systemctl restart kubelet",
		}
		commands := append(commands1, append(commandsDependent, commands2...)...)
		return commands, nil
	} else if centos, _ := regexp.MatchString("ID=\"centos\".*|ID=centos.*", string(output[:])); centos {
		// The commands including kubernetes & docker installation for CentOS, and also kubeadm join command
		commands := []string{
			"yum install yum-utils -y",
			"yum install epel-release -y",
			"yum update -y",
			"modprobe br_netfilter",
			"cat <<EOF > /etc/sysctl.d/k8s.conf",
			"net.bridge.bridge-nf-call-ip6tables = 1",
			"net.bridge.bridge-nf-call-iptables = 1",
			"EOF",
			"sysctl --system",
			"swapoff -a",
			"sed -e '/swap/ s/^#*/#/' -i /etc/fstab",
			"cat <<EOF > /etc/yum.repos.d/kubernetes.repo",
			"[kubernetes]",
			"name=Kubernetes",
			"baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\\$basearch",
			"enabled=1",
			"gpgcheck=1",
			"repo_gpgcheck=1",
			"gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg",
			"exclude=kubelet kubeadm kubectl",
			"EOF",
			"setenforce 0",
			"sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config",
			fmt.Sprintf("yum install docker kubeadm-%[1]s-0 kubectl-%[1]s-0 kubelet-%[1]s-0 kubernetes-cni -y --disableexcludes=kubernetes", kubernetesVersion),
			"systemctl enable --now kubelet",
			fmt.Sprintf("hostname %s", hostname),
			"systemctl enable docker",
			"systemctl start docker",
			node.CreateJoinToken("600s", hostname),
			"systemctl daemon-reload",
			"systemctl restart kubelet",
		}
		return commands, nil
	}
	return nil, fmt.Errorf("unknown")
}

// getUninstallCommands prepares the commands necessary according to the OS
func getUninstallCommands(conn *ssh.Client) ([]string, error) {
	sess, err := startSession(conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer sess.Close()
	// Detect the node OS
	output, err := sess.Output("cat /etc/os-release")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if ubuntuOrDebian, _ := regexp.MatchString("ID=\"ubuntu\".*|ID=ubuntu.*|ID=\"debian\".*|ID=debian.*", string(output[:])); ubuntuOrDebian {
		// The commands including kubeadm reset command, and kubernetes & docker installation for Ubuntu
		commands := []string{
			"kubeadm reset -f",
			"apt-get purge kubeadm kubectl kubelet kubernetes-cni kube* docker-engine docker docker.io docker-ce -y",
			"apt-get autoremove -y",
			"rm -rf ~/.kube",
			"iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X",
		}
		return commands, nil
	} else if centos, _ := regexp.MatchString("ID=\"centos\".*|ID=centos.*", string(output[:])); centos {
		// The commands including kubeadm reset command, and kubernetes & docker installation for CentOS
		commands := []string{
			"kubeadm reset -f",
			"yum remove kubeadm kubectl kubelet kubernetes-cni kube* docker docker-ce docker-ce-cli docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-engine -y",
			"yum clean all -y",
			"yum autoremove -y",
			"rm -rf ~/.kube",
			"iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X",
		}
		return commands, nil
	}
	return nil, fmt.Errorf("unknown")
}

// getReconfigurationCommands prepares the commands necessary according to the OS
func getReconfigurationCommands(conn *ssh.Client, hostname string) ([]string, error) {
	sess, err := startSession(conn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer sess.Close()
	// Detect the node OS
	output, err := sess.Output("cat /etc/os-release")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if ubuntuOrDebian, _ := regexp.MatchString("ID=\"ubuntu\".*|ID=ubuntu.*|ID=\"debian\".*|ID=debian.*", string(output[:])); ubuntuOrDebian {
		// The commands to set the hostname, restart docker & kubernetes and flush iptables on Ubuntu
		commands := []string{
			fmt.Sprintf("hostname %s", hostname),
			"systemctl stop docker",
			"systemctl stop kubelet",
			"iptables --flush",
			"iptables -tnat --flush",
			"systemctl start docker",
			"systemctl start kubelet",
		}
		return commands, nil
	} else if centos, _ := regexp.MatchString("ID=\"centos\".*|ID=centos.*", string(output[:])); centos {
		// The commands to set the hostname, restart docker & kubernetes and flush iptables on CentOS
		commands := []string{
			fmt.Sprintf("hostname %s", hostname),
			"systemctl stop docker",
			"systemctl stop kubelet",
			"iptables -F",
			"iptables -tnat -F",
			"systemctl start docker",
			"systemctl start kubelet",
		}
		return commands, nil
	}
	return nil, fmt.Errorf("unknown")
}
