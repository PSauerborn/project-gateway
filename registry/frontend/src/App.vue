<template>
  <v-app>
    <v-main>
      <v-container fluid>
        <v-row align="center" justify="center">
            <Registry />
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
import Registry from './components/Registry';

export default {
  name: 'App',

  components: {
    Registry,
  },

  data: () => ({
    //
  }),
};
</script>
