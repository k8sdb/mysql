# Copyright AppsCode Inc. and Contributors
#
# Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM mysql:5.7.25

COPY on-start.sh /
COPY peer-finder /usr/local/bin/

RUN chmod +x /on-start.sh

# For standalone mysql
# default entrypoint of parent mysql:5.7.25
# ENTRYPOINT ["docker-entrypoint.sh"]

# For mysql group replication
# ENTRYPOINT ["peer-finder"]
