<template>
    <v-col cols=8>
        <v-card style="padding-bottom:30px" flat="true">
            <v-row align="center" justify="center" class="title-row">
                <v-card-title>
                    Project Gateway Application Registry
                </v-card-title>
            </v-row>
            <v-row align="center" justify="center" class="title-row" style="margin:0px;">
                <v-card-subtitle>
                    Select an Application Below
                </v-card-subtitle>
            </v-row>
            <v-row align="center" justify="center">
                <application-entry v-for="(application, index) in applications" :application="application" :key="index"/>
            </v-row>
        </v-card>
    </v-col>
</template>

<script>

import ApplicationEntry from './ApplicationEntry'

export default {
    name: "Registry",
    components: {
        ApplicationEntry
    },
    methods: {

    },
    data() {
        return {
            applications: [
                {
                    application_name: 'Monty',
                    application_description: 'task management and analysis tool. create and manage daily tasks and gain insights on how to optimize your daily task flow',
                    application_id: '',
                    redirect_url: 'https://main.project-gateway.app/monty',
                    icon: 'mdi-file-tree'

                },
                {
                    application_name: 'Go Timesheets',
                    application_description: 'timesheet management tool used to track and analyse daily working hours',
                    application_id: '',
                    redirect_url: 'https://main.project-gateway.app/go-timesheets',
                    icon: 'mdi-briefcase-clock'

                }
            ]
        }
    }
}

</script>

<style scoped>

.title-row {
    margin-top: 50px;
}

</style>