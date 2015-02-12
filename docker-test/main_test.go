package main

import (
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	d "github.com/fsouza/go-dockerclient"
	"github.com/kr/pretty"
)

var (
	dockerHost     = os.Getenv("DOCKER_HOST")
	dockerTLS      = os.Getenv("DOCKER_TLS_VERIFY")
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
)

func newClient(t *testing.T) *d.Client {
	if dockerHost == "" {
		t.Fatal("No DOCKER_HOST")
	}

	if dockerTLS != "" {
		return newClientTLS(t)
	}

	c, err := d.NewClient(dockerHost)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newClientTLS(t *testing.T) *d.Client {
	if dockerCertPath == "" {
		t.Fatalf("DOCKER_TLS_VERIFY=%s but no DOCKER_CERT_PATH set", dockerTLS)
	}

	cert := dockerCertPath + "/cert.pem"
	key := dockerCertPath + "/key.pem"
	ca := dockerCertPath + "/ca.pem"

	c, err := d.NewTLSClient(dockerHost, cert, key, ca)
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func startContainer(t *testing.T, client *d.Client, image string) *d.Container {
	c, err := client.CreateContainer(d.CreateContainerOptions{
		Config: &d.Config{Image: image},
	})
	if err != nil {
		t.Fatal(err)
	}

	hostConfig := &d.HostConfig{PublishAllPorts: true}
	if err := client.StartContainer(c.ID, hostConfig); err != nil {
		t.Fatal(err)
	}

	info, err := client.InspectContainer(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	for !info.State.Running {
		i, err := client.InspectContainer(c.ID)
		if err != nil {
			t.Fatal(err)
		}
		info = i
	}
	time.Sleep(10 * time.Millisecond)

	return info
}

func redisHost(t *testing.T, c *d.Container) string {
	b, ok := c.NetworkSettings.Ports["6379/tcp"]
	if !ok {
		t.Logf("No port mapping to 6379 for %s", c.Name)
		t.Logf("ports: %# v", pretty.Formatter(c.NetworkSettings.Ports))
		t.Fail()
	}

	bind := b[0]
	u, err := url.Parse(dockerHost)

	if bind.HostIP == "0.0.0.0" && err == nil {
		h, _, err := net.SplitHostPort(u.Host)
		if err == nil {
			bind.HostIP = h
		}
	}

	return bind.HostIP + ":" + bind.HostPort
}

func getRedis(t *testing.T, c *d.Container) *Redis {
	r, err := newRedis(redisHost(t, c))
	if err != nil {
		t.Fatalf("error connecting to redis: %s", err)
	}

	return r
}

// func TestNewRedis(t *testing.T) {
// 	t.Parallel()

// 	client := newClient(t)
// 	c := startContainer(t, client, "redis")
// 	defer client.RemoveContainer(d.RemoveContainerOptions{ID: c.ID, Force: true})

// 	getRedis(t, c)
// }

// func TestPing(t *testing.T) {
// 	t.Parallel()

// 	client := newClient(t)
// 	c := startContainer(t, client, "redis")
// 	defer client.RemoveContainer(d.RemoveContainerOptions{ID: c.ID, Force: true})

// 	r := getRedis(t, c)

// 	if err := r.ping(); err != nil {
// 		t.Fatalf("error pinging db: %s", err)
// 	}
// }

func TestMe(t *testing.T) {
	c, err := NewTLSClient(dockerHost, dockerCertPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Run("scratch", ""); err != nil {
		t.Fatal(err)
	}
}
