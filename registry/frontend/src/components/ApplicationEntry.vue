<template>
    <v-sheet>
        <v-card @click.stop="redirect">
            <v-row align="center" justify="center" dense>
                <v-col cols=12 align="center" justify="center">
                    <v-card-text>
                        <v-icon :size="200">{{ application.icon }}</v-icon>
                    </v-card-text>
                </v-col>
            </v-row>
            <v-divider class="mx-4"></v-divider>
            <v-row align="center" justify="center" dense>
                <v-col cols=12 align="center" justify="center" dense>
                    <v-card-title>
                        {{ application.application_name }}
                    </v-card-title>
                </v-col>
                <v-col cols=12 align="center" justify="center" dense>
                    <v-card-subtitle>
                        {{ application.application_description }}
                    </v-card-subtitle>
                </v-col>
            </v-row>
        </v-card>
    </v-sheet>
</template>

<script>
export default {
    name: "ApplicationEntry",
    props: {
        application: {
            type: Object
        }
    },
    methods: {
        redirect: function() {
            window.location.href = this.application.redirect_url;
        }
    }
}
</script>