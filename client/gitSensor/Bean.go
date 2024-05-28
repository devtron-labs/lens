/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gitSensor

import "encoding/json"

type GitSensorResponse struct {
	Code   int                  `json:"code,omitempty"`
	Status string               `json:"status,omitempty"`
	Result json.RawMessage      `json:"result,omitempty"`
	Errors []*GitSensorApiError `json:"errors,omitempty"`
}

type GitSensorApiError struct {
	HttpStatusCode    int    `json:"-"`
	Code              string `json:"code,omitempty"`
	InternalMessage   string `json:"internalMessage,omitempty"`
	UserMessage       string `json:"userMessage,omitempty"`
	UserDetailMessage string `json:"userDetailMessage,omitempty"`
}

type GitChanges struct {
	Commits   []*Commit
	FileStats FileStats
}
type FileStats []FileStat

// FileStat stores the status of changes in content of a file.
type FileStat struct {
	Name     string
	Addition int
	Deletion int
}

type ReleaseChangesRequest struct {
	PipelineMaterialId int    `json:"pipelineMaterialId"`
	OldCommit          string `json:"oldCommit"`
	NewCommit          string `json:"newCommit"`
}
