from mythic_payloadtype_container.MythicCommandBase import *
import json


class JobKillArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = []

    async def parse_arguments(self):
        pass


class JobKillCommand(CommandBase):
    cmd = "jobkill"
    needs_admin = False
    help_cmd = "jobkill SOME-GUID-GOES-HERE"
    description = "Kill a job with the specified ID (from jobs command) - not all jobs are killable though."
    version = 1
    author = "@xorrior"
    argument_class = JobKillArguments
    attackmapping = []

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
