<template>
    <div class="bg-gray-50 p-6 rounded-lg shadow-inner">
        <h2 class="text-2xl font-bold text-gray-800 mb-4">Edit Story</h2>
        <div class="mb-4 text-left">
            <label for="storyTitle" class="block text-sm font-medium text-gray-700 mb-1">Title:</label>
            <input
                id="storyTitle"
                type="text"
                v-model="story.title"
                placeholder="Enter story title"
                class="focus:ring-indigo-500 focus:border-indigo-500"
            />
        </div>
        <div class="mb-6 text-left">
            <label for="storyDescription" class="block text-sm font-medium text-gray-700 mb-1">Description:</label>
            <textarea
                id="storyDescription"
                v-model="story.description"
                placeholder="Enter story description"
                rows="4"
                class="focus:ring-indigo-500 focus:border-indigo-500"
            ></textarea>
        </div>
        <div v-if="startLocation" class="mb-6 text-left">
            <h4 class="text-lg font-medium text-gray-800 mb-2">Starting Location: {{ startLocation.title }}</h4>
        </div>

        <!-- Events Section -->
        <div class="mt-6">
            <h3 class="text-xl font-semibold text-gray-800 mb-3">Events</h3>
            <button
                @click="createNewEvent"
                class="mb-4 px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-opacity-50">
                Create New Event
            </button>
            <div v-if="currentEvent" class="mt-4 p-4 border border-gray-300 rounded-md bg-white">
                 <TriggerActionEditor :triggerAction="currentEvent" :availableLocations="content.locations" @save-trigger-action="handleSaveEvent" @cancel-edit="handleCancelEventEdit" />
            </div>
            <ul v-if="storyEvents && storyEvents.length" class="list-disc pl-5 mb-4">
                <li v-for="(event, index) in storyEvents" :key="event.id || index" class="mb-2 flex justify-between items-center">
                    <span>Event {{ index + 1 }} <span v-if="event.condition && event.condition.comp" class="text-xs text-gray-500 ml-2">({{event.condition.comp.keyOne}} {{event.condition.comp.operation}} {{event.condition.comp.keyTwo}})</span></span>
                    <button
                        @click="editEvent(event)"
                        class="ml-2 px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 text-sm"
                    >
                        Edit
                    </button>
                </li>
            </ul>
            <p v-else class="text-gray-500 mb-4">No global events created yet.</p>
        </div>

        <div class="mt-6">
            <h3 class="text-xl font-semibold text-gray-800 mb-3">Locations</h3>
            <button
                @click="createNewLocation"
                class="mb-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50">
                Create New Location
            </button>
            <div v-if="currentLocation" class="mt-4 p-4 border border-gray-300 rounded-md bg-white">
                <h4 class="text-lg font-medium text-gray-800 mb-2">Edit Location: {{ currentLocation.title }}</h4>
                <label :for="'locationTitle-' + currentLocation.id" class="block text-sm font-medium text-gray-700 mb-1">Title:</label>
                <input
                    :id="'locationTitle-' + currentLocation.id"
                    type="text"
                    v-model="currentLocation.title"
                    placeholder="Enter location title"
                    class="focus:ring-indigo-500 focus:border-indigo-500 w-full"
                />
                <textarea
                    :id="'locationContent-' + currentLocation.id"
                    v-model="currentLocation.content"
                    placeholder="Enter location description"
                    rows="4"
                    class="focus:ring-indigo-500 focus:border-indigo-500 w-full"
                ></textarea>
            </div>
            <ul v-if="content.locations && content.locations.length" class="list-disc pl-5 mb-4">
                <li v-for="location in content.locations" :key="location.id" class="mb-2 flex justify-between items-center">
                    <span>{{ location.title }}</span>
                    <button
                        @click="editLocation(location)"
                        class="ml-2 px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 text-sm"
                    >
                        Edit
                    </button>
                    <button
                        @click="setStartingLocation(location)"
                        class="ml-2 px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 text-sm"
                    >
                        Set as Start
                    </button>
                </li>
            </ul>
            <p v-else class="text-gray-500 mb-4">No locations created yet.</p>
        </div>

        <!-- Actions Section -->
        <div class="mt-6">
            <h3 class="text-xl font-semibold text-gray-800 mb-3">Actions</h3>
            <button
                @click="createNewAction"
                class="mb-4 px-4 py-2 bg-purple-500 text-white rounded hover:bg-purple-600 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-opacity-50">
                Create New Action
            </button>
            <div v-if="currentAction" class="mt-4 p-4 border border-gray-300 rounded-md bg-white">
                <ActionEditor
                    :action="currentAction"
                    :availableLocations="content.locations"
                    @save-action="handleSaveAction"
                    @cancel-edit="handleCancelActionEdit" />
            </div>
            <ul v-if="content.actions && content.actions.length" class="list-disc pl-5 mb-4">
                <li v-for="(action, index) in content.actions" :key="action.id || index" class="mb-2 flex justify-between items-center">
                    <span>{{ action.title }}</span>
                    <button
                        @click="editAction(action)"
                        class="ml-2 px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 text-sm"
                    >
                        Edit
                    </button>
                </li>
            </ul>
            <p v-else class="text-gray-500 mb-4">No actions created yet.</p>
        </div>

        <button @click="saveChanges" class="w-full mt-8 px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-50">
            Save Story Changes
        </button>
        <p v-if="message" :class="['mt-4 text-sm', messageType === 'success' ? 'text-green-600' : 'text-red-600']">
            {{ message }}
        </p>
    </div>
</template>

<script>
import TriggerActionEditor from './TriggerActionEditor.vue';
import ActionEditor from './ActionEditor.vue';

export default {
    data() {
        // Initialize story data from window.initialStoryData if available, otherwise use defaults.
        const initialStory = window.initialStoryData || {
            title: 'New Story Title',
            description: 'Story introduction.',
            startLocationId: '',
            events: [],
        };
        let initialContent = window.initialContentData || {
            locations: [],
            actions: [],
        };

        // Ensure locations and actions are initialized if not present.
        if (!initialContent.locations) {
            initialContent.locations = [];
        }
        if (!initialContent.actions) {
            initialContent.actions = [];
        }

        let sloc = null;
        if (initialStory.startLocationId && initialContent.locations) {
            for (const loc of initialContent.locations) {
                if (loc.id == initialStory.startLocationId) {
                    sloc = loc;
                    break;
                }
            }
        }

        // Ensure storyEvents is initialized.
        const storyEvents = initialStory.events || [];

        return {
            story: {
                id: initialStory.id,
                title: initialStory.title,
                description: initialStory.description,
                startLocationId: initialStory.startLocationId,
            },
            content: initialContent,
            currentLocation: initialContent.locations.length > 0 ? initialContent.locations[0] : null,
            startLocation: sloc,
            storyEvents: storyEvents, // Global events for the story.
            currentEvent: null, // For editing a specific global event.
            currentAction: null, // For editing a specific action.
            message: '',
            messageType: ''
        };
    },
    components: {
      TriggerActionEditor,
      ActionEditor,
    },
    methods: {
        initializeNewEvent() {
            return {
                id: crypto.randomUUID(), // Temporary frontend ID
                condition: null,
                effects: [],
                isFinal: false,
            };
        },
        createNewEvent() {
            const newEvent = this.initializeNewEvent();
            if (!this.storyEvents) { // Should be initialized in data, but defensive check.
                this.storyEvents = [];
            }
            this.storyEvents.push(newEvent);
            this.currentEvent = newEvent;
            this.message = 'New global event ready for configuration.';
            this.messageType = '';
        },
        editEvent(event) {
            this.currentEvent = event;
        },
        handleSaveEvent(updatedEvent) {
            const index = this.storyEvents.findIndex(e => e.id === updatedEvent.id || (this.currentEvent && e === this.currentEvent));
            if (index !== -1) {
                this.storyEvents.splice(index, 1, updatedEvent);
            } else {
                this.storyEvents.push(updatedEvent);
            }
            this.currentEvent = null; // Close editor
            this.message = 'Global event saved.';
            this.messageType = 'success';
        },
        handleCancelEventEdit() {
            // Close the editor.
            this.currentEvent = null;
        },

        // Location Methods
        createNewLocation() {
            const newLocation = {
                id: crypto.randomUUID(),
                title: "Default Location Title",
                content: "Description goes here",
            };
	    this.content.locations.push(newLocation);
            this.currentLocation = newLocation;
            this.message = 'New location created.';
            this.messageType = '';
        },
        editLocation(location) {
	    this.currentLocation = location;
        },

        setStartingLocation(location) {
            this.story.startLocationId = location.id;
            this.startLocation = location;
        },

        // Action Methods
        initializeNewAction() {
            return {
                id: crypto.randomUUID(),
                title: 'New Action Title',
                description: 'Action description.',
                triggers: [],
            };
        },
        createNewAction() {
            const newAction = this.initializeNewAction();
            // Do not add to list yet, only set as current for editing
            this.currentAction = newAction;
            this.message = 'New action ready for configuration.';
            this.messageType = '';
        },
        editAction(action) {
            this.currentAction = JSON.parse(JSON.stringify(action)); // Edit a copy
            this.message = '';
        },
        handleSaveAction(updatedAction) {
            const index = this.content.actions.findIndex(a => a.id === updatedAction.id);
            if (index !== -1) {
                this.content.actions.splice(index, 1, updatedAction);
            } else {
                this.content.actions.push(updatedAction);
            }
            this.currentAction = null; // Close editor
            this.message = 'Action saved.';
            this.messageType = 'success';
        },
        handleCancelActionEdit() {
            this.currentAction = null;
        },

        // Save to backend.
        async saveChanges() {
            this.message = 'Saving...';
            this.messageType = '';
            try {
                const response = await fetch('/api/story/update', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                      story: { ...this.story, events: this.storyEvents },
                      content: this.content,
                    }),
                });

                if (response.ok) {
                    const result = await response.json();
                    console.log('Save successful:', result);
                    this.message = 'Changes saved successfully!';
                    this.messageType = 'success';
                    this.story = result.story
                    this.content = result.content
                } else {
                    const errorText = await response.text();
                    console.error('Failed to save changes:', response.status, errorText);
                    this.message = `Failed to save changes: ${response.status} ${errorText}`;
                    this.messageType = 'error';
                }
            } catch (error) {
                console.error('Error during save:', error);
                this.message = `An error occurred: ${error.message}`;
                this.messageType = 'error';
            }
        },
    },
};
</script>

<style scoped>
/* Scoped styles for this component can go here if needed */
/* For this example, we rely mostly on Tailwind classes */
</style>
