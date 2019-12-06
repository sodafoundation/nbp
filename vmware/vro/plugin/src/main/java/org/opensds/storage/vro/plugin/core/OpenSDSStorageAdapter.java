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

import javax.security.auth.login.LoginException;

import org.apache.log4j.Logger;

import ch.dunes.vso.sdk.api.IPluginAdaptor;
import ch.dunes.vso.sdk.api.IPluginEventPublisher;
import ch.dunes.vso.sdk.api.IPluginFactory;
import ch.dunes.vso.sdk.api.IPluginNotificationHandler;
import ch.dunes.vso.sdk.api.IPluginPublisher;
import ch.dunes.vso.sdk.api.PluginLicense;
import ch.dunes.vso.sdk.api.PluginLicenseException;
import ch.dunes.vso.sdk.api.PluginWatcher;

public class OpenSDSStorageAdapter implements IPluginAdaptor {
	static String pluginName;

	private static final OpenSDSStorageWatchersManager watchersManager = new OpenSDSStorageWatchersManager();

	private static final Logger log = Logger.getLogger(OpenSDSStorageAdapter.class);

	private OpenSDSStorageFactory factory;

	/**
	 * Create a new plugin factory. You could use the username / password to create
	 * a factory for each user. in this case, we will always use the same as the
	 * adapter is not user dependent.
	 *
	 * @param sessionID           is the session ID asking for the element. Sessions
	 *                            are created for each UI and for each user
	 *                            executing workflows
	 * @param username            is the user requesting the objects.
	 * @param password            password associated with the username
	 * @param notificationHandler is used by the plugin to notify the UI an element
	 *                            has changed, was deleted, etc.
	 * @return The factory from which the server will ask for Finders
	 */
	public IPluginFactory createPluginFactory(String sessionID, String username, String password,
			IPluginNotificationHandler notificationHandler)
			throws SecurityException, LoginException, PluginLicenseException {
		if (factory == null) {
			factory = new OpenSDSStorageFactory(notificationHandler);
		}
		return factory;
	}

	/**
	 * Allow to check the license. Not used any more
	 */
	@Deprecated
	public void installLicenses(PluginLicense[] licenses) throws PluginLicenseException {
		// Not used with the current license system.
	}

	/**
	 * Allow to register event publisher used by the policy Engine Not used in this
	 * sample
	 */
	public void registerEventPublisher(String type, String id, IPluginEventPublisher publisher) {
		getEventGenerator().addPolicyElement(type, id, publisher);
	}

	/**
	 * Get the plugin name from the Orchestrator server
	 */
	public void setPluginName(String pluginName) {
		OpenSDSStorageAdapter.pluginName = pluginName;
	}

	/**
	 * used by the server to destroy a previously created factory. If you need to do
	 * some specific cleaning when the factory is release, put the code from here.
	 * Not used in this sample
	 */
	public void uninstallPluginFactory(IPluginFactory plugin) {
	}

	/**
	 * Allow to unregister event publisher used by the policy Engine Not used in
	 * this sample
	 */
	public void unregisterEventPublisher(String type, String id, IPluginEventPublisher publisher) {
		getEventGenerator().removePolicyElement(type, id, publisher);
	}

	private OpenSDSStorageEventGenerator getEventGenerator() {
		return OpenSDSStorageEventGenerator.NEW_OPEN_SDS_STORAGE_EVENT_GENERATOR;
	}

	/**
	 * Used for long running Workflow
	 */
	public void addWatcher(PluginWatcher watcher) {
		if (log.isInfoEnabled()) {
			log.info("Adding watcher: {}" + watcher + "'");
		}
		watchersManager.addWatcher(watcher);
	}

	/**
	 * Used for long running Workflow
	 */
	public void removeWatcher(String watcherId) {
		if (log.isInfoEnabled()) {
			log.info("Removing watcher '" + watcherId + "'");
		}
		watchersManager.removeWatcher(watcherId);
	}

	/**
	 *
	 * @param pluginPublisher pluginPublisher
	 */
	public void setPluginPublisher(IPluginPublisher pluginPublisher) {
		watchersManager.setPluginPublisher(pluginPublisher);
	}
}
