from mythic_payloadtype_container.MythicCommandBase import *
import json


class ListtasksArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = []

    async def parse_arguments(self):
        pass


class ListtasksCommand(CommandBase):
    cmd = "listtasks"
    needs_admin = True
    help_cmd = "listtasks"
    description = "Obtain a list of processes with obtainable task ports on macOS. This command should be used to determine target processes for the libinject command"
    version = 1
    author = "@xorrior"
    argument_class = ListtasksArguments
    attributes = CommandAttributes(
        # uncomment when poseidon can dynamically compile commands
        supported_os=[SupportedOS.MacOS]
    )
    attackmapping = ["T1057"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        if task.callback.integrity_level <= 2:
            raise Exception("Error: the listtasks command requires elevated privileges")
        else:
            return task

    async def process_response(self, response: AgentResponse):
        pass
