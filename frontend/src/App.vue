<script>
import {Accounts, CurrentAccount, Peers, Self, SetExitNode, SwitchTo} from "../wailsjs/go/main/App";
import {EventsOn, EventsOnce} from "../wailsjs/runtime";

export default {
  data() {
    return {
      account: '',
      other_accounts: [],
      peers: [],
      self: {},
      selected_peer: null,
    }
  },
  methods: {
    load: async function() {
      this.account = await CurrentAccount();
      this.other_accounts = await Accounts();
      this.self = await Self();
      this.peers = await Peers();
      if (this.selected_peer === null) {
        this.selected_peer = this.self;
      }
    },
    switchAccount: async function(event) {
      const name = event.target.text;
      console.log("switch to", name);
      await SwitchTo(name)
      await this.load()
    },
    setExitNode: async function(event) {
      console.log("setting exit node")
      event.target.disabled = true;
      EventsOnce('exit_node_connect', () => {
        event.target.disabled = false;
      })
      await SetExitNode(this.selected_peer.dns_name);
    }
  },
  mounted() {
    this.load();
    EventsOn('update_all', () => this.load())
  },
  unmounted() {
    clearInterval(this.timer)
  }
}
</script>

<template>
  <div class="flex h-screen">
    <div class="w-1/3 h-full border-left-solid border-2 border-l-0 border-t-0 border-b-0 border-sky-500 overflow-scroll">
      <div class="py-4 px-3 rounded">
        <ul v-if="self != null" class="list-disc space-y-2">
          <li>
            <div class="flex-1 min-w-0">
              <a href="#" @click="selected_peer = self" class="flex items-center p-2 text-base font-normal text-gray-900 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700">
                <p class="text-md font-medium text-gray-900 truncate dark:text-white">
                  This machine
                </p>
              </a>
            </div>
          </li>
        </ul>
        <ul class="list-disc pt-5 space-y-2">
          <li v-for="namespace in peers">
            <p class="text-md font-medium text-gray-900 truncate dark:text-white">
              <span class="ml-3">{{ namespace.name }}</span>
            </p>
            <ol class="pl-5 mt-2 space-y-1 list-inside">
              <li v-for="peer in namespace.peers">
                <div class="flex items-center space-x-0">
                  <div class="flex-1 min-w-0">
                    <a href="#" @click="selected_peer = peer" class="flex items-center p-2 text-base font-normal text-gray-900 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700">
                      <p class="text-sm font-medium text-gray-900 truncate dark:text-white">
                        {{ peer.name }}
                      </p>
                    </a>
                  </div>
                  <div class="inline-flex items-center text-base font-semibold text-gray-900 dark:text-white">
                    <!-- https://flowbite.com/docs/components/badge/ -->
                    <span v-if="peer.exit_node" class="bg-red-100 text-red-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-red-200 dark:text-red-800">
                      Exit node
                    </span>
                    <span v-else-if="peer.exit_node_option" class="bg-yellow-100 text-yellow-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-yellow-200 dark:text-yellow-800">
                      Exit node
                    </span>
                  </div>
                  <div class="inline-flex items-center text-base font-semibold text-gray-900 dark:text-white">
                    <!-- https://flowbite.com/docs/components/badge/ -->
                    <span v-if="peer.online" class="bg-green-100 text-green-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-green-200 dark:text-green-800">
                      Online
                    </span>
                  </div>
                  <div class="inline-flex items-center text-base font-semibold text-gray-900 dark:text-white">
                    <!-- https://flowbite.com/docs/components/badge/ -->
                    <span class="bg-blue-100 text-blue-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-blue-200 dark:text-blue-800">
                      {{ peer.os }}
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
      <div v-if="selected_peer != null" class="flex flex-col mt-20 justify-center items-center px-2">
        <div>
          <h2 class="mt-8 text-center text-2xl font-bold text-zinc-100 cursor-default select-none"> {{ selected_peer.name }} </h2>
          <div class=" text-sm text-zinc-300">
            <p class="text-center"> {{ selected_peer.dns_name }} </p>
          </div>
        </div>
        <div class="w-full sm:max-w-md px-6 mt-2 mb-8">
          <div class="mt-8 space-y-8">
            <div class="mt-8 space-y-8 pl-4 pr-4 pt-4 pb-4 rounded border border-black-200 dark:border-white-700">
              <div v-for="ip in selected_peer.ips" class="flex items-center justify-between">
                <p class="text-center text-md font-medium text-gray-900 truncate dark:text-white">
                  {{ ip }}
                </p>
                <button type="button" class="text-blue-700 border border-blue-700 bg-white hover:bg-blue-700 hover:text-white focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm p-2.5 text-center inline-flex items-center mr-2 dark:border-blue-500 dark:text-blue-500 dark:hover:text-white dark:focus:ring-blue-800">
                  <svg height="16" viewBox="0 0 16 16" width="16" xmlns="http://www.w3.org/2000/svg"><path d="m0 3c0-1.644531 1.355469-3 3-3h5c1.644531 0 3 1.355469 3 3 0 .550781-.449219 1-1 1s-1-.449219-1-1c0-.570312-.429688-1-1-1h-5c-.570312 0-1 .429688-1 1v5c0 .570312.429688 1 1 1 .550781 0 1 .449219 1 1s-.449219 1-1 1c-1.644531 0-3-1.355469-3-3zm5 5c0-1.644531 1.355469-3 3-3h5c1.644531 0 3 1.355469 3 3v5c0 1.644531-1.355469 3-3 3h-5c-1.644531 0-3-1.355469-3-3zm2 0v5c0 .570312.429688 1 1 1h5c.570312 0 1-.429688 1-1v-5c0-.570312-.429688-1-1-1h-5c-.570312 0-1 .429688-1 1zm0 0" fill="#2e3436"/></svg>
                  <span class="sr-only">Icon description</span>
                </button>
              </div>
            </div>
            <div v-if="selected_peer !== self" class="flex items-center justify-between">
              <p class="text-center text-md font-medium text-gray-900 truncate dark:text-white">
                Use exit node
              </p>
              <label class="inline-flex relative items-center cursor-pointer">
                <input v-if="!selected_peer.online || !selected_peer.exit_node_option" @click="setExitNode" type="checkbox" value="" class="sr-only peer" disabled>
                <input v-else-if="selected_peer.exit_node" @click="setExitNode" type="checkbox" value="" class="sr-only peer" checked>
                <input v-else @click="setExitNode" type="checkbox" value="" class="sr-only peer">
                <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { ChevronDownIcon } from '@heroicons/vue/20/solid'
</script>

<style scoped></style>