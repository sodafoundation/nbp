// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package org.opensds.storage.vro.plugin.core;

import java.util.Hashtable;
import java.util.List;
import java.util.Map;
import java.util.Properties;
import java.util.Vector;

import org.apache.log4j.Logger;

import org.opensds.storage.vro.plugin.core.OpenSDSStorageEventGenerator.OpenSDSEventListener;

import ch.dunes.vso.sdk.api.IPluginPublisher;
import ch.dunes.vso.sdk.api.PluginWatcher;

public class OpenSDSStorageWatchersManager implements OpenSDSEventListener {
	private static final Logger log = Logger.getLogger(OpenSDSStorageWatchersManager.class);

	private final Map<String, PluginWatcher> watchers = new Hashtable<String, PluginWatcher>();

	private IPluginPublisher pluginPublisher;

	public OpenSDSStorageWatchersManager() {
		OpenSDSStorageEventGenerator.NEW_OPEN_SDS_STORAGE_EVENT_GENERATOR.addUniqueEventListener(this);
	}

	/**
	 * add watcher
	 * 
	 * @param watcher plugin watcher
	 */
	public void addWatcher(PluginWatcher watcher) {
		synchronized (watchers) {
			watchers.put(watcher.getId(), watcher);
		}
	}

	/**
	 * remove watcher
	 * 
	 * @param watcherId plugin watcher
	 */
	public void removeWatcher(String watcherId) {
		synchronized (watchers) {
			watchers.remove(watcherId);
		}
	}

	public void setPluginPublisher(IPluginPublisher pluginPublisher) {
		this.pluginPublisher = pluginPublisher;
	}

	/**
	 * event
	 * 
	 * @param starid    event id
	 *
	 * @param magnitude magnitude
	 */
	public void event(String starid, double magnitude) {
		synchronized (watchers) {
			List<String> watchersToRemove = new Vector<String>();
			for (PluginWatcher watcher : watchers.values()) {
				Properties props = watcher.getTrigger().getProperties();
				String wStarId = props.getProperty(OpenSDSStorageTriggerGenerator.STAR_ID);
				String wMagnitude = props.getProperty(OpenSDSStorageTriggerGenerator.MAGNITUDE);
				if (wStarId != null && wStarId.equals(starid)) {
					double wMagnLimit = Double.parseDouble(wMagnitude);
					if (magnitude >= wMagnLimit) {
						if (log.isInfoEnabled()) {
							log.info("pushWatcherEvent() for id '" + watcher.getId() + "'");
						}
						pluginPublisher.pushWatcherEvent(watcher.getId(), null);
						watchersToRemove.add(watcher.getId());
					}
				}
			}
			// Remove all treated watchers
			for (String toRemove : watchersToRemove) {
				watchers.remove(toRemove);
			}
		}
	}
}
