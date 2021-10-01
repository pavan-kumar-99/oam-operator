package aws

import (
	"fmt"

	aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	session "github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

var sess *s3.S3

func GetAuth() {
	sess = s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-south-1"),
	})))
}

func CreateS3(name string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	_, err := sess.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
				return aerr
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
				return aerr
			default:
				fmt.Println(aerr.Error())
				return aerr
			}
		} else {

			fmt.Println(err.Error())
		}
		return err
	}
	logrus.WithFields(logrus.Fields{
		"S3BucketName": name,
	}).Info("S3 bucket has been created")
	return nil
}

func DeleteS3(name string) error {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}
	_, err := sess.DeleteBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
				return aerr
			}
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
	logrus.WithFields(logrus.Fields{
		"S3BucketName": name,
	}).Info("S3 bucket has been deleted")
	return nil
}

func ListS3(name string) bool {
	input := &s3.ListBucketsInput{}
	result, err := sess.ListBuckets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return false
	}
	fmt.Println(result.Buckets)
	for _, bucket := range result.Buckets {
		if *bucket.Name == name {
			logrus.WithFields(logrus.Fields{
				"S3BucketName": name,
			}).Info("S3 bucket has been found")
			return true
		}
	}
	return false
}
