# Copyright AppsCode Inc. and Contributors
#
# Licensed under the AppsCode Community License 1.0.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM mysql:8.0.14

RUN set -x \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    netcat \
  && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/*

COPY osm /usr/local/bin/osm
COPY mysql-tools.sh /usr/local/bin/mysql-tools.sh

ENTRYPOINT ["mysql-tools.sh"]
