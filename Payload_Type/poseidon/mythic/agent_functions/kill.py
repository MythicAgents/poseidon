from mythic_payloadtype_container.MythicCommandBase import *
import json


class KillArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {}

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("Must supply a PID to kill")
        else:
            try:
                test = int(self.command_line)
            except:
                raise Exception("Must supply an integer parameter")
        pass


class KillCommand(CommandBase):
    cmd = "kill"
    needs_admin = False
    help_cmd = "kill [pid]"
    description = "Kill a process specified by PID."
    version = 1
    author = "@xorrior"
    argument_class = KillArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
