<script>
import {Accounts, CurrentAccount, SwitchTo} from "../wailsjs/go/main/App";

export default {
  data() {
    return {
      account: '',
      other_accounts: [],
    }
  },
  methods: {
    load: async function() {
      this.account = await CurrentAccount();
      this.other_accounts = await Accounts();
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
  <div class="h-full">
    <div class="flex flex-col h-full justify-start items-end px-8 py-2">
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
</template>

<script setup>
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { ChevronDownIcon } from '@heroicons/vue/20/solid'
</script>

<style scoped></style>