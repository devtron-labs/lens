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

create table if not exists app_release
(
    id                          serial PRIMARY KEY,
    app_id                      int not null,
    environment_id              int not null,
    ci_artifact_id              int not null,
    release_id                  int not null,
    pipeline_override_id        int not null,
    change_size_line_added      int not null default 0,
    change_size_line_deleted    int not null default 0,
    trigger_time                timestamptz not null,
    release_type                int not null,
    release_status              int not null,
    process_status              int not null,
    created_time                timestamptz not null,
    updated_time                timestamptz not null
);

create table if not exists lead_time
(
    id                          serial primary key,
    app_release_id              int not null references app_release,
    pipeline_material_id        int not null,
    commit_hash                 varchar(250) not null,
    commit_time                 timestamptz not null,
    lead_time                   bigint not null
);

create table if not exists pipeline_material
(
    app_release_id              int not null references app_release,
    pipeline_material_id        int not null,
    commit_hash                 varchar(250) not null
);


