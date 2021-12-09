package client

import "net/rpc"

func GetClient() (*rpc.Client,error){
	client, err := rpc.Dial("tcp","localhost:8011")
	if err != nil {
		return nil,err
	}
	return client,nil
}