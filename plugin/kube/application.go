package kube

import (
	"fmt"
	"github.com/baetyl/baetyl-cloud/common"
	"github.com/baetyl/baetyl-cloud/models"
	"github.com/baetyl/baetyl-cloud/plugin/kube/apis/cloud/v1alpha1"
	"github.com/baetyl/baetyl-go/log"
	specV1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/jinzhu/copier"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func toAppModel(app *v1alpha1.Application) *specV1.Application {
	description, _ := app.Annotations[common.AnnotationDescription]
	res := &specV1.Application{
		Name:        app.ObjectMeta.Name,
		Namespace:   app.ObjectMeta.Namespace,
		Version:     app.ObjectMeta.ResourceVersion,
		Description: description,
		Labels:      app.ObjectMeta.Labels,
	}

	err := copier.Copy(res, &app.Spec)
	if err != nil {
		panic(fmt.Sprintf("copier exception: %s", err.Error()))
	}
	res.CreationTimestamp = app.CreationTimestamp.Time.UTC()
	return res
}

func toAppListModel(list *v1alpha1.ApplicationList) *models.ApplicationList {
	res := &models.ApplicationList{
		Items: make([]models.AppItem, 0),
	}
	for _, item := range list.Items {
		description, _ := item.Annotations[common.AnnotationDescription]
		res.Items = append(res.Items, models.AppItem{
			Name:              item.ObjectMeta.Name,
			Type:              item.Spec.Type,
			Namespace:         item.ObjectMeta.Namespace,
			Version:           item.ObjectMeta.ResourceVersion,
			Labels:            item.ObjectMeta.Labels,
			Selector:          item.Spec.Selector,
			CreationTimestamp: item.CreationTimestamp.Time.UTC(),
			Description:       description,
			System:            item.Spec.System,
		})
	}

	res.Total = len(list.Items)
	return res
}

func fromAppModel(namespace string, app *specV1.Application) *v1alpha1.Application {
	res := &v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            app.Name,
			Namespace:       namespace,
			ResourceVersion: app.Version,
			Labels:          app.Labels,
			Annotations:     map[string]string{},
		},
	}

	if app.Description != "" {
		res.Annotations[common.AnnotationDescription] = app.Description
	}

	err := copier.Copy(&res.Spec, app)
	if err != nil {
		panic(fmt.Sprintf("copier exception: %s", err.Error()))
	}
	return res
}

func fromListOptionsModel(listOptions *models.ListOptions) *metav1.ListOptions {
	res := &metav1.ListOptions{}
	err := copier.Copy(res, listOptions)
	if err != nil {
		panic(fmt.Sprintf("copier exception: %s", err.Error()))
	}
	return res
}

func (c *client) GetApplication(namespace, name, version string) (*specV1.Application, error) {
	options := metav1.GetOptions{ResourceVersion: version}
	beforeRequest := time.Now().UnixNano()
	app, err := c.customClient.CloudV1alpha1().Applications(namespace).Get(name, options)
	afterRequest := time.Now().UnixNano()
	log.L().Debug("kube GetApplication", log.Any("cost time (ns)", afterRequest-beforeRequest))
	if err != nil {
		return nil, err
	}
	return toAppModel(app), nil
}

func (c *client) CreateApplication(namespace string, application *specV1.Application) (*specV1.Application, error) {
	app := fromAppModel(namespace, application)
	beforeRequest := time.Now().UnixNano()
	app, err := c.customClient.CloudV1alpha1().Applications(namespace).Create(app)
	afterRequest := time.Now().UnixNano()
	log.L().Debug("kube CreateApplication", log.Any("cost time (ns)", afterRequest-beforeRequest))
	if err != nil {
		return nil, err
	}
	res := toAppModel(app)
	return res, nil
}

func (c *client) UpdateApplication(namespace string, application *specV1.Application) (*specV1.Application, error) {
	app := fromAppModel(namespace, application)
	beforeRequest := time.Now().UnixNano()
	app, err := c.customClient.CloudV1alpha1().Applications(namespace).Update(app)
	afterRequest := time.Now().UnixNano()
	log.L().Debug("kube UpdateApplication", log.Any("cost time (ns)", afterRequest-beforeRequest))
	if err != nil {
		return nil, err
	}
	return toAppModel(app), nil
}

func (c *client) DeleteApplication(namespace, name string) error {
	beforeRequest := time.Now().UnixNano()
	err := c.customClient.CloudV1alpha1().Applications(namespace).Delete(name, &metav1.DeleteOptions{})
	afterRequest := time.Now().UnixNano()
	log.L().Debug("kube DeleteApplication", log.Any("cost time (ns)", afterRequest-beforeRequest))
	return err
}

func (c *client) ListApplication(namespace string, listOptions *models.ListOptions) (*models.ApplicationList, error) {
	beforeRequest := time.Now().UnixNano()
	list, err := c.customClient.CloudV1alpha1().Applications(namespace).List(*fromListOptionsModel(listOptions))
	afterRequest := time.Now().UnixNano()
	log.L().Debug("kube ListApplication", log.Any("cost time (ns)", afterRequest-beforeRequest))
	listOptions.Continue = list.Continue
	if err != nil {
		return nil, err
	}
	res := toAppListModel(list)
	res.ListOptions = listOptions
	return res, err
}
