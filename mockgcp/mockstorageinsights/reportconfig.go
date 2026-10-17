// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +tool:mockgcp-support
// proto.service: google.cloud.storageinsights.v1.StorageInsights
// proto.message: google.cloud.storageinsights.v1.ReportConfig

package mockstorageinsights

import (
	"context"
	"time"

	"github.com/GoogleCloudPlatform/k8s-config-connector/mockgcp/pkg/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "cloud.google.com/go/storageinsights/apiv1/storageinsightspb"
)

func (s *storageInsightsServer) GetReportConfig(ctx context.Context, req *pb.GetReportConfigRequest) (*pb.ReportConfig, error) {
	name, err := s.parseReportConfigName(req.GetName())
	if err != nil {
		return nil, err
	}

	fqn := name.String()

	obj := &pb.ReportConfig{}
	if err := s.storage.Get(ctx, fqn, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (s *storageInsightsServer) CreateReportConfig(ctx context.Context, req *pb.CreateReportConfigRequest) (*pb.ReportConfig, error) {
	reqReportConfigID := uuid.New().String()
	reqName := req.GetParent() + "/reportConfigs/" + reqReportConfigID
	name, err := s.parseReportConfigName(reqName)
	if err != nil {
		return nil, err
	}

	fqn := name.String()

	now := time.Now()

	obj := proto.Clone(req.GetReportConfig()).(*pb.ReportConfig)
	obj.Name = fqn
	obj.CreateTime = timestamppb.New(now)
	obj.UpdateTime = timestamppb.New(now)
	s.populateDefaultsForReportConfig(name, obj)

	if err := s.storage.Create(ctx, fqn, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (s *storageInsightsServer) populateDefaultsForReportConfig(name *reportConfigName, obj *pb.ReportConfig) {
	if obj.GetCsvOptions() != nil {
		if obj.GetCsvOptions().GetDelimiter() == "" {
			obj.GetCsvOptions().Delimiter = ","
		}
		if obj.GetCsvOptions().GetRecordSeparator() == "" {
			obj.GetCsvOptions().RecordSeparator = "\n"
		}
	}
}

func (s *storageInsightsServer) UpdateReportConfig(ctx context.Context, req *pb.UpdateReportConfigRequest) (*pb.ReportConfig, error) {
	reqName := req.GetReportConfig().GetName()
	name, err := s.parseReportConfigName(reqName)
	if err != nil {
		return nil, err
	}

	fqn := name.String()
	obj := &pb.ReportConfig{}
	if err := s.storage.Get(ctx, fqn, obj); err != nil {
		return nil, err
	}

	// Update fields based on updateMask
	mask := req.GetUpdateMask()
	if mask == nil || len(mask.GetPaths()) == 0 {
		// Replace entire object except output only fields
		cloned := proto.Clone(req.GetReportConfig()).(*pb.ReportConfig)
		cloned.Name = fqn
		cloned.CreateTime = obj.CreateTime
		cloned.UpdateTime = timestamppb.New(time.Now())
		s.populateDefaultsForReportConfig(name, cloned)
		if err := s.storage.Update(ctx, fqn, cloned); err != nil {
			return nil, err
		}
		return cloned, nil
	}

	updated := proto.Clone(obj).(*pb.ReportConfig)
	src := req.GetReportConfig()

	for _, path := range mask.GetPaths() {
		switch path {
		case "display_name", "displayName":
			updated.DisplayName = src.GetDisplayName()
		case "labels":
			updated.Labels = src.GetLabels()
		case "frequency_options", "frequencyOptions":
			updated.FrequencyOptions = src.GetFrequencyOptions()
		case "csv_options", "csvOptions":
			updated.ReportFormat = src.GetReportFormat()
		case "parquet_options", "parquetOptions":
			updated.ReportFormat = src.GetReportFormat()
		case "object_metadata_report_options", "objectMetadataReportOptions":
			updated.ReportKind = src.GetReportKind()
		case "object_metadata_report_options.metadata_fields", "objectMetadataReportOptions.metadataFields":
			if updated.GetObjectMetadataReportOptions() == nil {
				updated.ReportKind = &pb.ReportConfig_ObjectMetadataReportOptions{
					ObjectMetadataReportOptions: &pb.ObjectMetadataReportOptions{},
				}
			}
			updated.GetObjectMetadataReportOptions().MetadataFields = src.GetObjectMetadataReportOptions().GetMetadataFields()
		case "object_metadata_report_options.storage_destination_options.destination_path", "objectMetadataReportOptions.storageDestinationOptions.destinationPath":
			if updated.GetObjectMetadataReportOptions() == nil {
				updated.ReportKind = &pb.ReportConfig_ObjectMetadataReportOptions{
					ObjectMetadataReportOptions: &pb.ObjectMetadataReportOptions{},
				}
			}
			if updated.GetObjectMetadataReportOptions().GetStorageDestinationOptions() == nil {
				updated.GetObjectMetadataReportOptions().DestinationOptions = &pb.ObjectMetadataReportOptions_StorageDestinationOptions{
					StorageDestinationOptions: &pb.CloudStorageDestinationOptions{},
				}
			}
			updated.GetObjectMetadataReportOptions().GetStorageDestinationOptions().DestinationPath = src.GetObjectMetadataReportOptions().GetStorageDestinationOptions().GetDestinationPath()
		case "object_metadata_report_options.storage_destination_options.bucket", "objectMetadataReportOptions.storageDestinationOptions.bucket":
			if updated.GetObjectMetadataReportOptions() == nil {
				updated.ReportKind = &pb.ReportConfig_ObjectMetadataReportOptions{
					ObjectMetadataReportOptions: &pb.ObjectMetadataReportOptions{},
				}
			}
			if updated.GetObjectMetadataReportOptions().GetStorageDestinationOptions() == nil {
				updated.GetObjectMetadataReportOptions().DestinationOptions = &pb.ObjectMetadataReportOptions_StorageDestinationOptions{
					StorageDestinationOptions: &pb.CloudStorageDestinationOptions{},
				}
			}
			updated.GetObjectMetadataReportOptions().GetStorageDestinationOptions().Bucket = src.GetObjectMetadataReportOptions().GetStorageDestinationOptions().GetBucket()
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unsupported update mask path %q", path)
		}
	}

	updated.UpdateTime = timestamppb.New(time.Now())
	s.populateDefaultsForReportConfig(name, updated)

	if err := s.storage.Update(ctx, fqn, updated); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *storageInsightsServer) DeleteReportConfig(ctx context.Context, req *pb.DeleteReportConfigRequest) (*emptypb.Empty, error) {
	name, err := s.parseReportConfigName(req.GetName())
	if err != nil {
		return nil, err
	}

	fqn := name.String()

	deleted := &pb.ReportConfig{}
	if err := s.storage.Delete(ctx, fqn, deleted); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *storageInsightsServer) ListReportConfigs(ctx context.Context, req *pb.ListReportConfigsRequest) (*pb.ListReportConfigsResponse, error) {
	response := &pb.ListReportConfigsResponse{}
	findKey := req.GetParent() + "/reportConfigs"

	if err := s.storage.List(ctx, (&pb.ReportConfig{}).ProtoReflect().Descriptor(), storage.ListOptions{
		Prefix: findKey,
	}, func(obj proto.Message) error {
		item := obj.(*pb.ReportConfig)
		response.ReportConfigs = append(response.ReportConfigs, item)
		return nil
	}); err != nil {
		return nil, err
	}
	return response, nil
}
