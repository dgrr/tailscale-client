<script>
import {Accounts, CurrentAccount, Peers, SwitchTo} from "../wailsjs/go/main/App";

export default {
  data() {
    return {
      account: '',
      other_accounts: [],
      peers: [],
    }
  },
  methods: {
    load: async function() {
      this.account = await CurrentAccount();
      this.other_accounts = await Accounts();
      this.peers = await Peers();
      console.log(this.peers)
    },
    switchAccount: async function(event) {
      const name = event.target.text;
      console.log("switch to", name);
      await SwitchTo(name)
      await this.load()
    }
  },
  mounted() {
    this.load()
  }
}
</script>

<template>
  <div class="flex h-screen">
    <div class="w-1/3 h-full border-solid border-2 border-sky-500">
      <div class="overflow-y-auto py-4 px-3 rounded">
        <ul class="list-disc space-y-2">
          <li v-for="namespace in peers">
            <p class="text-md font-medium text-gray-900 truncate dark:text-white">
              <span class="ml-3">{{ namespace.name }}</span>
            </p>
            <ol class="pl-5 mt-2 space-y-1 list-inside">
              <li v-for="peer in namespace.peers">
                <div class="flex items-center space-x-3">
                  <div class="flex-1 min-w-0">
                    <a href="#" class="flex items-center p-2 text-base font-normal text-gray-900 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700">
                      <p class="text-sm font-medium text-gray-900 truncate dark:text-white">
                        {{ peer.name }}
                      </p>
                    </a>
                  </div>
                  <div class="inline-flex items-center text-base font-semibold text-gray-900 dark:text-white">
                    <!-- https://flowbite.com/docs/components/badge/ -->
                    <span v-if="peer.exit_node_option" class="bg-blue-100 text-blue-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-blue-200 dark:text-blue-800">
                      Exit node
                    </span>
                  </div>
                </div>
              </li>
            </ol>
          </li>
        </ul>
      </div>
    </div>
    <div class="w-2/3 h-full">
      <div class="flex flex-col h-25 justify-end items-end px-8 py-2">
        <Menu as="div" class="relative inline-block text-left">
          <MenuButton class="inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100">
            Accounts ({{ account }})
            <ChevronDownIcon class="-mr-1 ml-2 h-5 w-5" aria-hidden="true" />
          </MenuButton>
          <transition
              enter-active-class="transition duration-100 ease-out"
              enter-from-class="transform scale-95 opacity-0"
              enter-to-class="transform scale-100 opacity-100"
              leave-active-class="transition duration-75 ease-out"
              leave-from-class="transform scale-100 opacity-100"
              leave-to-class="transform scale-95 opacity-0"
          >
            <MenuItems class="absolute right-0 z-10 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
              <div class="py-1">
                <MenuItem v-slot="{ active }">
                  <a href="#" :class="[active ? 'bg-gray-100 text-gray-900' : 'text-gray-700', 'block px-4 py-2 text-sm']">Signed in as <br><b>{{ account }}</b></a>
                </MenuItem>
              </div>
              <div class="py-1">
                <MenuItem v-for="account in other_accounts" v-slot="{ active }">
                  <a href="#" @click="switchAccount" :class="[active ? 'bg-gray-100 text-gray-900' : 'text-gray-700', 'block px-4 py-2 text-sm']">{{ account }}</a>
                </MenuItem>
              </div>
              <div class="py-1">
                <MenuItem v-slot="{ active }">
                  <a href="#" :class="[active ? 'bg-gray-100 text-gray-900' : 'text-gray-700', 'block px-4 py-2 text-sm']">Logout</a>
                </MenuItem>
              </div>
            </MenuItems>
          </transition>
        </Menu>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { ChevronDownIcon } from '@heroicons/vue/20/solid'
</script>

<style scoped></style>