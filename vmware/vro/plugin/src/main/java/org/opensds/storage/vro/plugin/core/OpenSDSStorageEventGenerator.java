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
import java.util.Vector;

import org.apache.log4j.Logger;

import ch.dunes.vso.sdk.api.IPluginEventPublisher;
import ch.dunes.vso.sdk.api.IPluginFactory;

public class OpenSDSStorageEventGenerator {
	private static final Logger log = Logger.getLogger(OpenSDSStorageEventGenerator.class);
	/**
	 * OpenSDSStorage Event Generator Object
	 */
	public static final OpenSDSStorageEventGenerator NEW_OPEN_SDS_STORAGE_EVENT_GENERATOR = new OpenSDSStorageEventGenerator();

	private Hashtable<String, Vector<IPluginEventPublisher>> policyElements = new Hashtable<String, Vector<IPluginEventPublisher>>();

	private OpenSDSEventListener eventListener;

	/**
	 * OpenSDSStorage Event Generator Singleton Method
	 * 
	 * @param factory Plugin Factory
	 * @return OpenSDSStorageEventGenerator
	 */
	public static OpenSDSStorageEventGenerator createScriptingSingleton(IPluginFactory factory) {
		return NEW_OPEN_SDS_STORAGE_EVENT_GENERATOR;
	}

	/**
	 * add policy element
	 * 
	 * @param dskType   disk type
	 * @param id        disk id
	 * @param publisher publisher
	 */
	public void addPolicyElement(String dskType, String id, IPluginEventPublisher publisher) {
		String key = dskType + "' / '" + id;
		if (log.isInfoEnabled()) {
			log.info("Registering element to watch : '" + key + "'");
		}

		Vector<IPluginEventPublisher> publishers = policyElements.get(key);
		if (publishers == null) {
			publishers = new Vector<IPluginEventPublisher>();
			policyElements.put(key, publishers);
		}
		publishers.add(publisher);
	}

	/**
	 * remove policy element
	 * 
	 * @param dskType   disk type
	 * @param id        disk id
	 * @param publisher publisher
	 */
	public boolean removePolicyElement(String dskType, String id, IPluginEventPublisher publisher) {
		String key = dskType + "' / '" + id;
		if (log.isInfoEnabled()) {
			log.info("Unregistering element to watch : '" + key + "'");
		}

		Vector<IPluginEventPublisher> publishers = policyElements.get(key);
		if (publishers != null) {
			publishers.remove(publisher);
			if (publishers.size() == 0) {
				policyElements.remove(key);
			}
		}
		return (policyElements.size() > 0);
	}

	/**
	 * generate event
	 * 
	 * @param dskType   disk type
	 * @param id        event id
	 * @param magnitude magnitude
	 */
	public void generateEvent(String dskType, String id, double magnitude) {
		String key = dskType + "' / '" + id;
		if (log.isInfoEnabled()) {
			log.info("Generate Flare Event for : '" + key + "' with magnitude " + magnitude);
		}
		Vector<IPluginEventPublisher> publishers = policyElements.get(key);
		if (publishers != null) {
			for (IPluginEventPublisher publisher : publishers) {
				publisher.pushGauge(dskType, id, "Flare", "magnitude", magnitude);
			}
		}
		if (dskType.trim().length() == 0) {
			eventListener.event(id, magnitude);
		}
	}

	/**
	 * add unique event listener
	 * 
	 * @param eventListener event listerner
	 */
	public void addUniqueEventListener(OpenSDSEventListener eventListener) {
		this.eventListener = eventListener;
	}

	public interface OpenSDSEventListener {
		/**
		 * Event Listener Constructor
		 * 
		 * @param starid    event id
		 *
		 * @param magnitude magnitude
		 */
		void event(String starid, double magnitude);
	}
}
