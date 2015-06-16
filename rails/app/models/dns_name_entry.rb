# Copyright 2015, Greg Althaus
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#  http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

class DnsNameEntry < ActiveRecord::Base

  audited

  belongs_to :network_allocation
  belongs_to :dns_name_filter

  after_commit :on_create_hooks, on: :create
  after_commit :on_destroy_hooks, on: :destroy

  scope :for_filter, ->(dnf) { where(:dns_name_filter_id => dnf.id) }
  scope :for_network_allocation, ->(na) { where(:network_allocation_id => na.id) }

  def on_destroy_hooks
    Rails.logger.fatal("GREG: DNE - destroy hook: #{self.network_allocation.inspect}")
    BarclampDns::MgmtService.remove_ip_address(self)
    DnsNameFilter.claim_by_any(self.network_allocation.reload) if self.network_allocation
  end

  def on_create_hooks
    BarclampDns::MgmtService.add_ip_address(self)
  end

  def release
    self.network_allocation = nil
    destroy!
  end

end
