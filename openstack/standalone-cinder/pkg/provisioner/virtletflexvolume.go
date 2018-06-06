package provisioner

import (
	"errors"
	"strings"

	"fmt"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/openstack/standalone-cinder/pkg/volumeservice"
	"k8s.io/api/core/v1"
)

const (
	virtletFlexVolumeRBDType = "virtletFlexVolumeRBD"
	virtletDriver            = "virtlet/flexvolume_driver"
	virtletRbdType           = "ceph"
	secretNamespace          = "kube-system"
)

type flexVolumeRbdMapper struct {
	volumeMapper
	cb clusterBroker
}

func (m *flexVolumeRbdMapper) BuildPVSource(p *cinderProvisioner, conn volumeservice.VolumeConnection, options controller.VolumeOptions) (*v1.PersistentVolumeSource, error) {
	mons := getMonitors(conn)
	if mons == nil {
		return nil, errors.New("No monitors could be parsed from connection info")
	}
	splitName := strings.SplitN(conn.Data.Name, "/", 2)
	if len(splitName) != 2 {
		return nil, errors.New("Field 'name' cannot be split into pool and image")
	}

	secretName := getRbdSecretName(options.PVC)

	secretObj, err := m.cb.getSecret(p, secretNamespace, secretName)
	if err != nil {
		msg := fmt.Sprintf("Failed to get secret:%s namespace:%s err:%v ", secretName, secretNamespace, err)
		return nil, errors.New(msg)
	}

	secret := make(map[string]string)
	for name, data := range secretObj.Data {
		secret[name] = string(data)
	}

	option := map[string]string{}
	option["type"] = virtletRbdType
	option["monitor"] = mons[0]
	option["user"] = conn.Data.AuthUsername
	option["pool"] = splitName[0]
	option["volume"] = splitName[1]
	option["secret"] = secret["key"]
	return &v1.PersistentVolumeSource{
		FlexVolume: &v1.FlexPersistentVolumeSource{
			Driver:  virtletDriver,
			Options: option,
		},
	}, nil
}

func (m *flexVolumeRbdMapper) AuthSetup(p *cinderProvisioner, options controller.VolumeOptions, conn volumeservice.VolumeConnection) error {
	return nil
}

func (m *flexVolumeRbdMapper) AuthTeardown(p *cinderProvisioner, pv *v1.PersistentVolume) error {
	return nil
}
