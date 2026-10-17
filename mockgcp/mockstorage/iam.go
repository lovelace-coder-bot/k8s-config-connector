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

package mockstorage

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	pb "github.com/GoogleCloudPlatform/k8s-config-connector/mockgcp/generated/mockgcp/storage/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// serveBucketIAM handles GET and PUT /storage/v1/b/{bucket}/iam
func (s *MockService) serveBucketIAM(w http.ResponseWriter, r *http.Request) bool {
	// Pattern: /storage/v1/b/{bucket}/iam
	path := r.URL.Path
	if !strings.HasPrefix(path, "/storage/v1/b/") || !strings.HasSuffix(path, "/iam") {
		return false
	}

	parts := strings.Split(strings.TrimPrefix(path, "/storage/v1/b/"), "/")
	if len(parts) != 2 || parts[1] != "iam" {
		return false
	}
	bucketName := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.handleGetBucketIAM(w, r, bucketName)
		return true
	case http.MethodPut:
		s.handlePutBucketIAM(w, r, bucketName)
		return true
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return true
	}
}

func (s *MockService) handleGetBucketIAM(w http.ResponseWriter, r *http.Request, bucket string) {
	ctx := r.Context()
	fqn := "projects/_/buckets/" + bucket + "/iam"

	policy := &pb.Policy{}
	err := s.storage.Get(ctx, fqn, policy)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Return default empty policy
			version := int32(1)
			if r.URL.Query().Get("optionsRequestedPolicyVersion") == "3" {
				version = 3
			}
			policy = &pb.Policy{
				Kind:       PtrTo("storage#policy"),
				ResourceId: PtrTo("projects/_/buckets/" + bucket),
				Version:    PtrTo(version),
				Etag:       computePolicyEtag(policy),
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeIAMPolicyResponse(w, policy)
}

func (s *MockService) handlePutBucketIAM(w http.ResponseWriter, r *http.Request, bucket string) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	policy := &pb.Policy{}
	if err := json.Unmarshal(body, policy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	policy.Kind = PtrTo("storage#policy")
	policy.ResourceId = PtrTo("projects/_/buckets/" + bucket)
	if policy.Version == nil {
		policy.Version = PtrTo(int32(1))
	}
	policy.Etag = computePolicyEtag(policy)

	fqn := "projects/_/buckets/" + bucket + "/iam"
	if err := s.storage.Update(ctx, fqn, policy); err != nil {
		if status.Code(err) == codes.NotFound {
			if err := s.storage.Create(ctx, fqn, policy); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeIAMPolicyResponse(w, policy)
}

func writeIAMPolicyResponse(w http.ResponseWriter, policy *pb.Policy) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Server", "UploadServer")
	w.Header().Set("Vary", "Origin")
	w.Header().Add("Vary", "X-Origin")
	w.Header().Add("Vary", "Referer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Xss-Protection", "0")
	w.WriteHeader(http.StatusOK)

	bytes, err := json.Marshal(policy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func computePolicyEtag(policy *pb.Policy) []byte {
	clone := proto.Clone(policy).(*pb.Policy)
	clone.Etag = nil
	b, err := proto.Marshal(clone)
	if err != nil {
		panic(fmt.Sprintf("converting to proto: %v", err))
	}
	hash := md5.Sum(b)
	return hash[:]
}
