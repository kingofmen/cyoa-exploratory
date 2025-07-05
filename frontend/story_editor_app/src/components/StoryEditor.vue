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
        };

        return {
            story: {
	        id: initialStory.id,
                title: initialStory.title,
                description: initialStory.description,
            },
            message: '',
            messageType: ''
        };
    },
    methods: {
        async saveChanges() {
            this.message = 'Saving...';
            this.messageType = '';
            try {
                const response = await fetch('/api/story/update', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(this.story),
                });

                if (response.ok) {
                    const result = await response.json();
                    console.log('Save successful:', result);
                    this.message = 'Changes saved successfully!';
                    this.messageType = 'success';
		    this.story = result.story
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
