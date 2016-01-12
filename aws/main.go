package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
)

func GetVolumeId(i string, c *ec2.EC2) *string {

	opts := &ec2.DescribeInstancesInput{
		DryRun: aws.Bool(false),
		InstanceIds: []*string{
			aws.String(i),
		},
	}
	instanceStatus, err := c.DescribeInstances(opts)
	if err != nil {
		fmt.Println("Error Occured :", err.Error())
		os.Exit(1)
	}
	return instanceStatus.Reservations[0].Instances[0].BlockDeviceMappings[0].Ebs.VolumeId
}

func CreateSnapshot(v *string, i string, c *ec2.EC2) *string {
	params := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(*v), // Required
		Description: aws.String("Snapshot for :" + i),
		DryRun:      aws.Bool(false),
	}
	SnapshotStatus, err := c.CreateSnapshot(params)
	if err != nil {
		fmt.Println("Error Occured :", err.Error())
		os.Exit(1)
	}
	return SnapshotStatus.SnapshotId
}

func main() {
	// Place the creds under under ~/.aws/credentials file(recomended).
	// Export the secret keys to environment variables, including region.

	Config := &aws.Config{
		Region: aws.String("ap-southeast-1"),
	}
	svc := ec2.New(session.New(), Config)
	InstanceId := "i-81b93d0f"

	// Get the Volume Id
	VolumeId := GetVolumeId(InstanceId, svc)
	fmt.Println(*VolumeId)

	SnapshotId := CreateSnapshot(VolumeId, InstanceId, svc)
	fmt.Println(*SnapshotId)
}
