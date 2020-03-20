/*

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

package controllers

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openapiv1 "github.com/samze/crdotohttp/api/v1"
)

// RequestReconciler reconciles a Request object
type RequestReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=openapi.pivotal.io,resources=requests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openapi.pivotal.io,resources=requests/status,verbs=get;update;patch

func (r *RequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("request", req.NamespacedName)
	log.Info("reconcile")

	// fetch
	request := openapiv1.Request{}
	if err := r.Get(ctx, req.NamespacedName, &request); err != nil {
		log.Error(err, "unable to fetch Request")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("request info", "request:", request)

	//check status
	if request.Status.Code != "" {
		log.Info("request already made")
		return ctrl.Result{}, nil
	}

	httpReq := HTTPRequest{
		Path:    request.Spec.Path,
		Method:  request.Spec.Method,
		Body:    request.Spec.Body,
		Headers: request.Spec.Headers,
	}

	code, body, err := httpReq.Do()
	if err != nil {
		return ctrl.Result{}, err
	}

	request.Status.Code = code
	request.Status.Body = body

	if err := r.Status().Update(ctx, &request); err != nil {
		log.Error(err, "unable to update request status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

type HTTPRequest struct {
	Path    string
	Method  string
	Body    string
	Headers []string
}

func (r HTTPRequest) Do() (string, string, error) {
	client := &http.Client{}

	var requestBody io.ReadCloser
	if r.Body != "" {
		requestBody = ioutil.NopCloser(strings.NewReader(r.Body))
	}

	req, err := http.NewRequest(strings.ToUpper(r.Method), r.Path, requestBody)
	if err != nil {
		return "", "", err
	}

	for _, h := range r.Headers {
		split := strings.Split(h, ":")
		req.Header.Add(split[0], split[1])
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	code := strconv.Itoa(resp.StatusCode)

	return code, string(respBody), nil
}

func (r *RequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openapiv1.Request{}).
		Complete(r)
}
