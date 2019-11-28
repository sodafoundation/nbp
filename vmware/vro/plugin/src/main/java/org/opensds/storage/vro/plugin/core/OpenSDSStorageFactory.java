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

import java.util.List;
import java.util.Vector;

import org.opensds.storage.vro.plugin.adapter.opensds.services.Configuration;
import org.opensds.storage.vro.plugin.adapter.opensds.services.VolumeService;
import org.opensds.storage.vro.plugin.adapter.opensds.services.OpenSDSInfo;
import org.apache.log4j.Logger;

import org.opensds.storage.vro.plugin.adapter.opensds.OpenSDSStorageRepository;

import ch.dunes.vso.sdk.api.HasChildrenResult;
import ch.dunes.vso.sdk.api.IPluginFactory;
import ch.dunes.vso.sdk.api.IPluginNotificationHandler;
import ch.dunes.vso.sdk.api.PluginExecutionException;
import ch.dunes.vso.sdk.api.QueryResult;

public class OpenSDSStorageFactory implements IPluginFactory {
	private static final Logger log = Logger.getLogger(OpenSDSStorageFactory.class);

	private Configuration conf = new Configuration("conf", "conf");

	private VolumeService volumeService = new VolumeService();

	public OpenSDSStorageFactory(IPluginNotificationHandler notificationHandler) {
		super();
		new OpenSDSStorageEventListener(OpenSDSStorageRepository.getUniqueInstance(), notificationHandler);
	}

	/**
	 * Process a command directly transfered from the GUI.
	 * 
	 * @param cmd is totally plugin specific
	 */
	public void executePluginCommand(String cmd) throws PluginExecutionException {
		// No command supported right now.
	}

	/**
	 * Find an item by its ID (for a given type)
	 * 
	 * @param type is the type (as defined in the Finders part of the vso.xml file
	 * @param id   the unique id (for the given type) of the element to get
	 * @return The element if found (null if asked for the OpenSDS)
	 * @throws IndexOutOfBoundsException if the Type of the element is unknown.
	 */
	public Object find(String type, String id) {
		if (log.isDebugEnabled()) {
			log.debug("find: " + type + ", " + id);
		}
		/*
		 * if (type.equals("LUN")) { OpenSDSInfo OpenSDSInfo = OpenSDSStorageRepository
		 * .getUniqueInstance().getOpenSDSInfoById(id); return
		 * OpenSDSInfo.getLUNById(id); }
		 */

		if (type.equals("OpenSDSInfo")) {
			return OpenSDSStorageRepository.getUniqueInstance().getOpenSDSInfoById(id);
		}
		if (type.equals("VolumeService")) {
			return volumeService;
		}
		if (type.equals("Conf")) {
			return conf; // No object for OpenSDS defined now
		}

		log.error("Type " + type + " + unknown for OpenSDS Storage Plugin");
		return new Object();
	}

	/**
	 * Get all item of a given type (Finder type).
	 * 
	 * @param type  is the type of the elements to return
	 * @param query is the plugin specific query. Could be ignored...
	 * @return A QueryResult contains the resulting planet.
	 * @throws IndexOutOfBoundsException if the type is not a valid type
	 */
	public QueryResult findAll(String type, String query) {
		if (log.isDebugEnabled()) {
			log.debug("findAll: " + type + ", " + query);
		}
		List list = new Vector(); // The list can contains any element from the plugin
		if (type.equals("OpenSDSInfo")) {
			list = new Vector();
			list.addAll(OpenSDSStorageRepository.getUniqueInstance().getOpenSDSInfos());
		} else if (type.equals("OpenSDS")) {
			list = new Vector();
		} else {
			throw new IndexOutOfBoundsException("Type " + type + " unknown for plugin OpenSDS Storage");
		}
		return new QueryResult(list);
	}

	/**
	 * Find a list of elements related to another one (the parent).
	 * 
	 * @param parentType   is the Finder type (vso.xml) of the "parent" element
	 * @param parentId     is the ID of the "parent" element
	 * @param relationName is the name of the relation (vso.xml) linking the parent
	 *                     with the elements you want to return. Relation names are
	 *                     define in the vso.xml for each finders.
	 * @return The related element through the relation name.
	 * @throws IndexOutOfBoundsException if the relationName is not valid for a
	 *                                   given Finder type. Null if the "parent"
	 *                                   provided is not intended to have children.
	 *                                   (prevent crash, as we provide unknown for
	 *                                   hasChildrenInRelation
	 */
	public List findRelation(String parentType, String parentId, String relationName) {
		if (log.isDebugEnabled()) {
			log.debug("findRelation: " + parentType + ", " + parentId + ", " + relationName);
		}
		if (parentId == null) {
			if (relationName.equals("OpenSDSInfos")) {
				List<OpenSDSInfo> list = new Vector<OpenSDSInfo>();
				list.addAll(OpenSDSStorageRepository.getUniqueInstance().getOpenSDSInfos());
				return list;
			} else {
				return new Vector();
			}
		}
		return new Vector();
	}

	/**
	 * This method allow the UI (or Web Service clients) to know quickly if the
	 * "parent" has children for the given relation. In this sample, we don't check
	 * and juste return unknown.
	 * 
	 * @param parentType   is the Finder type (vso.xml) of the "parent" element
	 * @param parentId     is the ID of the "parent" element
	 * @param relationName is the name of the relation (vso.xml) for which you want
	 *                     to know the children lists.
	 * @return The related element through the relation name.
	 */
	public HasChildrenResult hasChildrenInRelation(String parentType, String parentId, String relationName) {
		return HasChildrenResult.Unknown; // don't bother with this for now.
	}

	/**
	 * Invalidate a defined Finder. Not used is this sample
	 */
	public void invalidate(String type, String id) {
		// We never need to invalidate, as we don't have cache
	}

	/**
	 * Invalidate all finders Not used in this sample
	 */
	public void invalidateAll() {
		// We never need to invalidate, as we don't have cache
	}

}
