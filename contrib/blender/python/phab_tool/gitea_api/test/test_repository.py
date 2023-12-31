"""
    Gitea API.

    This documentation describes the Gitea API.  # noqa: E501

    The version of the OpenAPI document: 1.19.0+dev-367-g8042ec1b9
    Generated by: https://openapi-generator.tech
"""


import sys
import unittest

import phab_tool.gitea_api
from phab_tool.gitea_api.model.external_tracker import ExternalTracker
from phab_tool.gitea_api.model.external_wiki import ExternalWiki
from phab_tool.gitea_api.model.internal_tracker import InternalTracker
from phab_tool.gitea_api.model.permission import Permission
from phab_tool.gitea_api.model.repo_transfer import RepoTransfer
from phab_tool.gitea_api.model.user import User
globals()['ExternalTracker'] = ExternalTracker
globals()['ExternalWiki'] = ExternalWiki
globals()['InternalTracker'] = InternalTracker
globals()['Permission'] = Permission
globals()['RepoTransfer'] = RepoTransfer
globals()['User'] = User
from phab_tool.gitea_api.model.repository import Repository


class TestRepository(unittest.TestCase):
    """Repository unit test stubs"""

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def testRepository(self):
        """Test Repository"""
        # FIXME: construct object with mandatory attributes with example values
        # model = Repository()  # noqa: E501
        pass


if __name__ == '__main__':
    unittest.main()
