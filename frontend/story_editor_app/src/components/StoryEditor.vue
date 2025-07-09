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
        <button @click="saveChanges" class="w-full">
            Save Changes
        </button>
        <p v-if="message" :class="['mt-4 text-sm', messageType === 'success' ? 'text-green-600' : 'text-red-600']">
            {{ message }}
        </p>
    </div>
</template>

<script>
export default {
    data() {
        // Initialize story data from window.initialStoryData if available, otherwise use defaults.
        const initialStory = window.initialStoryData || {
            title: 'New Story Title',
            description: 'Story introduction.',
	    startLocationId: '',
        };
	let initialContent = window.initialContentData || {
                locations: [],
	}

        // Ensure locations is initialized if not present.
        if (!initialContent.locations) {
            initialContent.locations = [];
        }

	let sloc = null;
	for (const loc of initialContent.locations) {
	  if (loc.id == initialStory.startLocationId) {
	    sloc = loc
	    break
	  }
	}

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
            message: '',
            messageType: ''
        };
    },
    methods: {
        createNewLocation() {
            const newLocation = {
                id: crypto.randomUUID(),
                title: "Default Location title",
		content: "Description goes here",
            };
            this.content.locations.push(newLocation);
            this.currentLocation = newLocation; // Select the new location for editing
            this.message = 'New location created';
            this.messageType = '';
        },
        editLocation(location) {
            this.currentLocation = location;
        },
        setStartingLocation(location) {
            this.story.startLocationId = location.id;
	    this.startLocation = location
        },
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
		      story: this.story,
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
