<style lang="scss">
  @import './main';

  .blur {
    filter: blur(5px);
  }
</style>

<template>
  <div class="bg-const d-flex flex-grow-1 flex-column">
    <nav
      v-if="!isAuthenticated"
      class="navbar sticky-top navbar-dark bg-darker shadow-sm"
    >
      <p class="lead text-white m-0 my-2">Tukija helps with accessing and exporting Youtube channel memberships.</p>
      <button @click="authenticate" class="btn btn-sm btn-purple">Login with Google</button>
    </nav>

    <main
      role="main"
      :class="{'blur':!isAuthenticated}"
      class="d-flex flex-grow-1 flex-column"
    >
      <div class="row no-gutters flex-grow-1 d-flex">
        <div class="col-12 col-sm-3 col-md-3 col-lg-3 col-xl-2
          bg-dark border-right border-purple
          flex-grow-1 d-flex flex-column
        ">
          <fieldset
            class="p-3"
            :disabled="!isAuthenticated"
          >
            <div class="form-group">
              <label class="text-light">Membership levels</label>
              <div class="input-group input-group-sm">
                <select class="custom-select" v-model="membershipLevel">
                  <option v-if="!isAuthenticated" disabled selected>Requires authentication...</option>
                  <option v-if="isAuthenticated && membershipLevels.length === 0" disabled selected>No membership levels...</option>
                  <option
                    :value="level.id"
                    v-for="level in membershipLevels"
                  >
                    {{level.snippet.levelDetails.displayName}}
                  </option>

                </select>
                <div class="input-group-append">
                  <button
                    @click="getMembershipLevels"
                    class="btn btn-sm btn-outline-secondary"
                  >
                    ⟳
                  </button>
                </div>
              </div>
              <small class="form-text text-muted">The minimum level a member should have.</small>
            </div>
            <button
              class="btn btn-sm btn-purple btn-block"
              @click="getMembers"
            >
              Get memberships
            </button>
          </fieldset>
          <!-- <div class="mt-auto p-3">
            <button
              @click="saveAsCsv"
              class="btn btn-primary btn-block"
              :disabled="!isAuthenticated"
            >
              Save as CSV
            </button>
          </div> -->
          <div class="border-top border-darker bg-darker p-3 mt-auto">
            <button
              v-if="isAuthenticated"
              @click="signout"
              class="btn btn-outline-secondary btn-block btn-sm"
            >
              Sign out
            </button>
          </div>
        </div>
        <div class="
          col-12 col-sm-9 col-md-9 col-lg-9 col-xl-10
          border-top border-right border-primary border-purple
          flex-column flex-grow-1 d-flex
        ">
          <div v-if="error" class="alert alert-danger mb-0 rounded-0 p-1">
            <h5>
              {{error.error}}
              <button class="btn btn-sm p-0 px-1 btn-outline-danger pull-right" @click="error = false">Dismiss</button>
            </h5>
            <p>{{error.message}}</p>
            <pre style="white-space: pre-wrap">{{error.stack}}</pre>
          </div>

          <nav class="navbar navbar-dark bg-purple text-light">
            <div class="d-flex flex-grow-1">
              </span>
                {{memberships.length || '-'}} Memberships
              </span>
              <div class="ml-auto">
                <button
                  @click="saveAsCsv"
                  class="btn btn-outline-light btn-sm"
                  :disabled="!isAuthenticated"
                >
                  Save as CSV
                </button>
              </div>
          </nav>

          <table
            v-if="isAuthenticated"
            class="table table-dark table-sm table-striped shadow-sm"
          >
            <thead>
              <tr class="font-weight-normal">
                <th scope="col">Display name</th>
                <th scope="col">Channel</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td v-if="!memberships" colspan="4">
                  <p class="text-muted text-center m-3">
                    No data. Select a membership level and then
                    <button
                      class="btn btn-sm btn-purple btn-inline"
                      @click="getMembers"
                      :disabled="!membershipLevel"
                    >
                      Get memberships
                    </button>
                  </p>
                </td>
              </tr>
              <tr v-for="member in memberships">
                <td>{{member.snippet.memberDetails.displayName}}</td>
                <td>
                  <a
                    :href="member.snippet.memberDetails.channelUrl"
                    target="_blank"
                    class="text-light"
                    @click.prevent="openExternal(member.snippet.memberDetails.channelUrl)"
                  >
                    {{member.snippet.memberDetails.channelId}}
                  </a>
                </td>
                <td class="text-right">
                  <button
                    class="btn btn-sm p-0 px-1 btn-outline-secondary"
                    @click="inspect(member)"
                  >
                    Inspect
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <div class="bg-darker text-info mt-auto">
            <pre class="m-3" v-if="debug">{{debug}}</pre>
          </div>
        </div>

      </div>
    </main>
  </div>
</template>

<script>
  import {ipcRenderer} from 'electron'

  const Emit = (topic, data = false) => {
    const now = Date.now();

    return new Promise((resolve, reject) => {
      ipcRenderer.send(topic, {now, data});
      ipcRenderer.on(`${topic}-${now}`, (event, res) => {
        if (res.error) {
          reject(res);
          return;
        }

        resolve(res);
      });
    })
  }

  export default {
    name: 'tukija',
    methods: {
      openExternal(url) {
        Emit('openExternal', url);
      },
      saveAsCsv() {
        Emit('save', this.membershipsToCsv()).then((yolo) => {
          console.log(yolo)
        })
      },
      inspect(member) {
        this.debug = member;
      },
      signout() {
        Emit('signout').then(() => {
          this.error = false;
          this.isAuthenticated = false;
        }, (error) => {
          this.error = error;
        });
      },
      authenticate() {
        Emit('authenticate').then(() => {
          this.isAuthenticated = true;
          this.getMembershipLevels();
        }, (error) => {
          this.error = error;
        });
      },
      getMembershipLevels() {
        if (!this.isAuthenticated) return;

        Emit('membershipLevels').then((data) => {
          this.debug = data;
          this.membershipLevels = data;
        }, (error) => {
          this.error = error;
        });
      },
      getMembers() {
        if (!this.isAuthenticated) return;
        Emit('members', this.membershipLevel).then((data) => {
          this.debug = data;
          this.memberships = data;
        }, (error) => {
          this.error = error;
        })
      },
      membershipsToCsv() {
        let csv = '"Display name"\r\n';

        this.memberships.forEach((member) => {
          csv += `"${member.snippet.memberDetails.displayName}"\r\n`;
        });

        return csv;
      }
    },
    mounted() {
      Emit('state:read').then((data) => {
        this.membershipLevel = data.membershipLevel;
      });

      Emit('isAuthenticated').then((data) => {
        this.isAuthenticated = data.isAuthenticated;

        if (this.isAuthenticated) {
          this.getMembershipLevels();
        }
      })
    },
    beforeDestroy() {
      Emit('state:write', {
        membershipLevel: this.membershipLevel
      });
    },
    data () {
      return {
        isAuthenticated: false,
        error: false,
        membershipLevels: [],
        membershipLevel: false,
        memberships: [],
        debug: false
      }
    }
  }
</script>
