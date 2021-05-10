from mythic_payloadtype_container.MythicCommandBase import *
import json


class SetEnvArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {}

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("No arguments given. Must be of format 'setenv NAME VALUE'")
        pass


class SetEnvCommand(CommandBase):
    cmd = "setenv"
    needs_admin = False
    help_cmd = "setenv [param] [value]"
    description = "Sets an environment variable to your choosing."
    version = 1
    author = "@xorrior"
    argument_class = SetEnvArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
